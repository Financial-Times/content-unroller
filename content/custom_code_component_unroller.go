package content

import (
	"errors"
	"fmt"
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
		// TODO make to Debugf or remove
		u.log.WithUUID(event.uuid).WithTransactionID(event.tid).Infof("No embedded components for CCC UUID: %v", event.uuid)
		return ccc, nil
	}

	if len(emContentUUIDs) == 0 {
		return ccc, nil
	}

	embedded := []Content{}
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
	// Take Care for Potential Inner CCC inside the bodyXML
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
	// embeddedComponent is expected to be inner CustomCodeComponent up to level2 depth or ImageSet with Image members.
	for _, embeddedComponent := range embedsElements {
		embedFoundMember, found := resolveContent(embeddedComponent, foundContent)
		if !found {
			u.log.WithUUID(uuid).WithTransactionID(tid).Debugf("Cannot match to any found content UUID: %v", embeddedComponent)
			// Specified UUID in the `id`
			foundContent[embeddedComponent] = Content{id: createID(u.apiHost, "content", embeddedComponent)}
			continue
		}

		localLog := u.log.WithUUID(uuid).WithTransactionID(tid)

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
			localLog.Debug("No embedded unrollable content inside the bodyXML")
			continue
		}

		// If these emContentUUIDs are not already got, get this content via Reader (one more REST call)
		// add the inner ImageSet and Images and inner CustomCodeComponent to an inner embeds node.
		innerEmbeds := []Content{}
		newInnerContent := []string{}
		for _, innerUUID := range emContentUUIDs {
			innerContentFound, found := resolveContent(innerUUID, foundContent)
			if found {
				// Expand members field of the new inner content, which was just loaded before append to embeds node
				expMembers, err := u.unrollMembersForImageSetInCCC(innerContentFound, foundContent, tid)
				if err != nil {
					localLog.Infof("failed to fill inner content members field. Check for ImageSet %s in CCC, which has not published images", innerUUID)
					// TODO what to do in this case - we have CCC, it has ImageSet, but the image members are not found!!!
					innerEmbeds = append(innerEmbeds, innerContentFound)
					continue
				}
				// replace potential ImageSet with the same ImageSet with unrolled members. Same element if no members.
				innerContentFound = expMembers
				innerEmbeds = append(innerEmbeds, innerContentFound)

			} else {
				newInnerContent = append(newInnerContent, innerUUID)
			}
		}
		if len(newInnerContent) > 0 {
			// Read Inner Content For Unrolling via reader.go->Get for new UUID(s)
			tempContentMap, err := u.reader.Get(newInnerContent, tid)
			if err != nil {
				// TODO replace with Debugf as we could have CCC with placeholder UUID for ImageSet or inner CCC.
				localLog.WithError(err).Infof("failed to read CustomCodeComponent inner content %s", err)
				continue
			}
			// Append Inner Content to embeds for new content from inner bodyXML, which was not present in parent bodyXML
			for _, newItemUUID := range newInnerContent {
				newInnerContentFound, found := resolveContent(newItemUUID, tempContentMap)
				if found {
					// Expand members field of the new inner content, which was just loaded before append to members node
					expMembers, err := u.unrollMembersForImageSetInCCC(newInnerContentFound, tempContentMap, tid)
					if err != nil {
						localLog.Infof("failed to fill inner content members field. Check for ImageSet %s in CCC, which has not published images", newItemUUID)
						// TODO what to do in this case - we have CCC, it has ImageSet, but the image members are not found!!!
						innerEmbeds = append(innerEmbeds, newInnerContentFound)
						continue
					}
					newInnerContentFound = expMembers
					innerEmbeds = append(innerEmbeds, newInnerContentFound)
				} else {
					// We have an embedded content in bodyXML, with UUID that is not present in known resources,
					// and it is not found in a separate get call to downstream service. We cannot unroll it.
					// TODO what to return in this case?!?
					// Skip the UUID in embeds collection?
					// Or add it as content with id element only?
					localLog.Info("failed to find inner for CustomCodeComponent content for unroll")
				}
			}
		}
		// Add embeds element to CustomCodeComponent similar to Article, so the internal content is unrolled as well
		embedFoundMember[embeds] = innerEmbeds
	}
}

func (u *UniversalUnroller) unrollMembersForImageSetInCCC(innerContent Content, loadedContent map[string]Content, tid string) (Content, error) {
	members, ok := innerContent[membersField].([]interface{})
	if !ok {
		// There are no members field in this supposed to be ImageSet, maybe it is something else.
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
		images, err = u.reader.Get(newImageUUIDs, tid)
		if err != nil {
			// TODO add logs that call to downstream service (content-public-read) has failed to get UUIDs.
			return nil, err
		}
	}

	unrolledImages := []Content{}
	for _, imageUUID := range imageUUIDs {
		newImage := loadedContent[imageUUID]
		if newImage == nil {
			newImage = images[imageUUID]
		}
		// TODO what if any of the images in the ImageSet is not found/loaded? Shall we add something in members?
		if newImage == nil {
			// TODO change to Debugf when ready
			u.log.Infof("Not Found Image: %s for ImageSet", imageUUID)
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
		unrollMembersContent, err := u.unrollMembersForImageSetInCCC(embedFoundMember, foundContent, tid)
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
