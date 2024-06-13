package content

import (
	"errors"
	"fmt"
	"github.com/Financial-Times/go-logger/v2"
)

// CustomCodeComponent does not have members field with its content. It has xmlBody with the inner content.
// The content could be an ImageSet with Images or embedded CustomCodeComponent, which could contain more
// ImageSet with images inside. So we need to process the bodyXML, then expand members of ImageSets.
// And if supported - to go recursively for inner CustomCodeComponents
func (u *UniversalUnroller) unrollCustomCodeComponent(event UnrollEvent) (Content, error) {
	if !validateCustomCodeComponent(event.c) {
		return nil, ErrValidating
	}

	ccc := event.c.clone()

	acceptedTypes := []string{ImageSetType, DynamicContentType, ClipSetType, CustomCodeComponentType}

	//embedded - image set(s), clip set(s), dynamic content(s) and custom code component(s) from ccc
	emContentUUIDs, foundEmbedded := extractEmbeddedContentByType(ccc, u.log, acceptedTypes, event.tid, event.uuid)
	if !foundEmbedded {
		u.log.WithUUID(event.uuid).WithTransactionID(event.tid).Debugf("No embedded components for CCC UUID: %v", event.uuid)
		return ccc, nil
	}

	if len(emContentUUIDs) == 0 {
		return ccc, nil
	}

	var embedded []Content
	// Read inner content for unrolling from CCC
	contentMap, err := u.reader.Get(emContentUUIDs, event.tid)
	if err != nil {
		// TODO shall we create empty ID in embeds if we cannot read the content by the provided UUID?
		// Or just return nil and the error.
		for _, innerUUID := range emContentUUIDs {
			embedded = append(embedded, Content{id: createID(u.apiHost, "content", innerUUID)})
		}
		ccc[embeds] = embedded
		return ccc, errors.Join(err, fmt.Errorf("error while getting expanded content for uuid: %v as uuid(s): %v", event.uuid, emContentUUIDs))
	}

	// Take Care for ImageSet to enroll members on root level in main bodyXML ft-content tag
	u.resolveModelsForSetsMembers(emContentUUIDs, contentMap, event.tid, event.uuid)
	// Take Care for Potential Inner CCC(s) and ImageSet(s) inside the bodyXML
	u.resolveModelsForInnerBodyXML(emContentUUIDs, contentMap, acceptedTypes, event.tid, event.uuid)

	// Add all unrolled embeds to the CCC
	if len(emContentUUIDs) > 0 {
		for _, emb := range emContentUUIDs {
			embedded = append(embedded, contentMap[emb])
		}
		ccc[embeds] = embedded
	}

	return ccc, nil
}

func validateCustomCodeComponent(c Content) bool {
	_, ok := c[bodyXMLField]

	return ok && checkType(c, CustomCodeComponentType)
}

func (u *UniversalUnroller) resolveModelsForInnerBodyXML(
	embedsElements []string,
	foundContent map[string]Content,
	acceptedTypes []string,
	tid string,
	uuid string) {
	// embeddedComponentUUID is expected to be inner CustomCodeComponent up to level2 depth or ImageSet with Image members.
	// Here we process inner CustomCodeComponent, but we have to process also the ImageSet in inner CCC if present.
	localLog := u.log.WithUUID(uuid).WithTransactionID(tid)

	for _, embeddedComponentUUID := range embedsElements {
		embedFoundMember, found := resolveContent(embeddedComponentUUID, foundContent)
		if !found {
			localLog.Debugf("cannot match to any found content UUID: %v", embeddedComponentUUID)
			// TODO shall we create empty ID in embeds if we cannot read the content by the provided UUID? Like row 36.
			// Specified UUID in the `id`
			foundContent[embeddedComponentUUID] = Content{id: createID(u.apiHost, "content", embeddedComponentUUID)}
			continue
		}

		// Custom Code Component does not have members field, but it has bodyXML,
		// so we need to run one more check for inner images and inner CCC
		// similar to the extractEmbeddedContentByType in UniversalUnroller service.
		// We need to just skip the member in case of errors.
		rawBody, foundBody := embedFoundMember[bodyXMLField]
		if !foundBody {
			// localLog.Debug("Missing body. Skipping expanding embedded components and images.")
			continue
		}

		bodyXML := rawBody.(string)
		emContentUUIDs, err := getEmbedded(u.log, bodyXML, acceptedTypes, tid, uuid)
		if err != nil {
			localLog.WithError(err).Errorf("Cannot parse bodyXML for CCC content %s", err.Error())
			continue
		}

		if len(emContentUUIDs) == 0 {
			localLog.Debug("No embedded unrollable content inside the CCC bodyXML")
			continue
		}

		// We have found CCC with BodyXML, which has <ft-content> tag for unrolling, so we process it.
		// If these emContentUUIDs are not already got, get this content via Reader (one more REST call)
		// add the inner ImageSet and Images and inner CustomCodeComponent to an inner embeds node.
		innerEmbeds, err := processContentForEmbeds(emContentUUIDs, foundContent, u.reader, u.log, tid, uuid)
		if err != nil {
			localLog.Infof("failed to load content to unroll in any of: %v", emContentUUIDs)
			continue
		}

		// Add embeds element to CustomCodeComponent similar to Article, so the internal content is unrolled as well
		embedFoundMember[embeds] = innerEmbeds
	}
}

func processContentForEmbeds(emContentUUIDs []string, foundContent map[string]Content, reader Reader, log *logger.UPPLogger, tid string, uuid string) (contentForEmbeds []Content, err error) {
	localLog := log.WithUUID(uuid).WithTransactionID(tid)
	var innerEmbeds []Content
	var newInnerContent []string
	// Get the UUIDs, which are not loaded yet. We may have some ImageSet, which was already loaded in parent Content
	for _, innerUUID := range emContentUUIDs {
		if foundContent[innerUUID] == nil {
			newInnerContent = append(newInnerContent, innerUUID)
		}
	}
	if len(newInnerContent) > 0 {
		// Read Inner Content For Unrolling via reader.go->Get for new UUID(s)
		tempContentMap, err := reader.Get(newInnerContent, tid)
		if err != nil {
			// TODO replace with Debugf as we could have CCC with placeholder UUID for ImageSet or inner CCC.
			localLog.WithError(err).Infof("failed to read CustomCodeComponent inner content %s", err)
			return nil, err
		}
		// And add the new content to already loaded content for unrolling
		for _, newInnerUUID := range newInnerContent {
			foundContent[newInnerUUID] = tempContentMap[newInnerUUID]
		}
	}
	for _, innerUUID := range emContentUUIDs {
		innerContentFound, found := resolveContent(innerUUID, foundContent)
		if found {
			// Expand members field of the new inner content, which was just loaded before append to embeds node
			expMembers, err := unrollMembersForImageSetInCCC(innerContentFound, foundContent, reader, log, tid)
			if err != nil {
				localLog.Infof("failed to fill inner content members field. Check for ImageSet %s in CCC, which has not published images", innerUUID)
				// TODO what to do in this case - we have CCC, it has ImageSet, but the image members are not found!!!
				innerEmbeds = append(innerEmbeds, innerContentFound)
				continue
			}
			// replace potential ImageSet with the same ImageSet with unrolled members. Same element if no members.
			innerEmbeds = append(innerEmbeds, expMembers)
		} else {
			// We have an embedded content in bodyXML, with UUID that is not present in known resources,
			// and it is not found in a separate get call to downstream service (content-public-read).
			// We cannot unroll it.
			// TODO what to return in this case?!?
			// Skip the UUID in embeds collection? (currently implemented)
			// Or add it as content with id element only?
			// TODO replace log to Debugf when ready the implementation
			localLog.Infof("failed to find inner for CustomCodeComponent content for unroll: %s", innerUUID)
		}
	}
	return innerEmbeds, nil
}

func unrollMembersForImageSetInCCC(innerContent Content, loadedContent map[string]Content, reader Reader, log *logger.UPPLogger, tid string) (Content, error) {
	members, ok := innerContent[membersField].([]interface{})
	if !ok {
		// There is no `members` field in this supposed to be ImageSet, maybe it is something else. Return as it is.
		return innerContent, nil
	}
	if len(members) == 0 {
		return innerContent, nil
	}

	var imageUUIDs []string
	for _, m := range members {
		// search for `id` fields in each member
		memberID, ok := m.(map[string]interface{})["id"].(string)
		if !ok {
			return nil, ErrConverting
		}
		// extract uuid part of the whole id, ignoring the prefix part
		uuid, err := extractUUIDFromString(memberID)
		if err != nil {
			// TODO add logs that the UUID cannot be extracted from id field
			return nil, err
		}
		imageUUIDs = append(imageUUIDs, uuid)
	}

	// Filter only the new images to get from the service
	var newImageUUIDs []string
	for _, graphicUUID := range imageUUIDs {
		if loadedContent[graphicUUID] == nil {
			newImageUUIDs = append(newImageUUIDs, graphicUUID)
		}
	}

	images := map[string]Content{}
	if len(newImageUUIDs) > 0 {
		var err error
		images, err = reader.Get(newImageUUIDs, tid)
		if err != nil {
			// TODO add logs that call to downstream service (content-public-read) has failed to get UUIDs.
			return nil, err
		}
		// add loaded images (or other members) to already loaded content
		// TODO what if the ImageSet and Image has one and the same UUID, but different type?!?
		for graphicUUID, graphicValue := range images {
			loadedContent[graphicUUID] = graphicValue
		}
	}

	unrolledImages := []Content{}
	for _, imageUUID := range imageUUIDs {
		newImage, ok := loadedContent[imageUUID]
		// TODO what if any of the images in the ImageSet is not found/loaded? Shall we add something in members as id?
		if !ok || newImage == nil {
			// TODO change to Debugf when ready
			log.Infof("Not Found Image: %s for ImageSet", imageUUID)
			continue
		}
		unrolledImages = append(unrolledImages, newImage)
	}

	returnContent := innerContent.clone()
	returnContent[membersField] = unrolledImages
	return returnContent, nil
}

func (u *UniversalUnroller) resolveModelsForSetsMembers(
	embedsElements []string,
	foundContent map[string]Content,
	tid string,
	uuid string) {
	for _, embeddedComponentUUID := range embedsElements {
		embedFoundMember, found := resolveContent(embeddedComponentUUID, foundContent)
		if !found {
			u.log.WithUUID(uuid).WithTransactionID(tid).Debugf("Cannot match to any found content UUID: %v", embeddedComponentUUID)
			// Specified UUID in the `id`
			foundContent[embeddedComponentUUID] = Content{id: createID(u.apiHost, "content", embeddedComponentUUID)}
			continue
		}
		unrollMembersContent, err := unrollMembersForImageSetInCCC(embedFoundMember, foundContent, u.reader, u.log, tid)
		if err != nil {
			localLog := u.log.WithUUID(uuid).WithTransactionID(tid)
			localLog.Infof("failed to fill inner content members field. Check for ImageSet %s in CCC, which has not published images", embeddedComponentUUID)
			// TODO what to do in this case - we have CCC, it has ImageSet, but the image members are not found!!!
			continue
		}
		// Replace Image Set with the same Image Set with unrolled members field.
		foundContent[embeddedComponentUUID] = unrollMembersContent
	}
}
