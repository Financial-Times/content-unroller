package content

import (
	"errors"
	"fmt"

	"github.com/Financial-Times/go-logger/v2"
)

type DefaultUnroller struct {
	reader  Reader
	log     *logger.UPPLogger
	apiHost string
}

func NewDefaultUnroller(r Reader, log *logger.UPPLogger, apiHost string) *DefaultUnroller {
	return &DefaultUnroller{
		reader:  r,
		log:     log,
		apiHost: apiHost,
	}
}

func (u *DefaultUnroller) Unroll(req UnrollEvent) (Content, error) {
	if !validateDefaultContent(req.c) {
		return req.c, ErrValidating
	}

	cc := req.c.clone()

	acceptedTypes := []string{ImageSetType, DynamicContentType, ClipSetType, CustomCodeComponentType}
	schema := u.createContentSchema(cc, acceptedTypes, req.tid, req.uuid)
	if schema == nil {
		return cc, nil
	}

	// Read content for unrolling
	contentMap, err := u.reader.Get(schema.toArray(), req.tid)
	if err != nil {
		return req.c, errors.Join(err, fmt.Errorf("error while getting expanded content for uuid: %v", req.uuid))
	}

	u.resolveModelsForSetsMembers(schema, contentMap, req.tid, req.uuid)
	u.resolveModelsForInnerBodyXML(schema, contentMap, acceptedTypes, req.tid, req.uuid)

	// Add unrolled mainImage
	mainImageUUID := schema.get(mainImageField)
	if mainImageUUID != "" {
		cc[mainImageField] = contentMap[mainImageUUID]
	}

	// Add all unrolled embeds
	embeddedContentUUIDs := schema.getAll(embeds)
	if len(embeddedContentUUIDs) > 0 {
		embedded := []Content{}
		for _, emb := range embeddedContentUUIDs {
			embedded = append(embedded, contentMap[emb])
		}
		cc[embeds] = embedded
	}

	// Add unrolled promotionalImage as alternativeImages.promotionalImage
	promImgUUID := schema.get(promotionalImage)
	if promImgUUID != "" {
		pi, found := contentMap[promImgUUID]
		if found {
			cc[altImagesField].(map[string]interface{})[promotionalImage] = pi
		}
	}

	return cc, nil
}

func (u *DefaultUnroller) createContentSchema(cc Content, acceptedTypes []string, tid string, uuid string) Schema {
	schema := make(Schema)

	localLog := u.log.WithUUID(uuid).WithTransactionID(tid)

	//mainImage as mainImage
	mainImageUUID, foundMainImg := extractMainImageContentByType(cc, u.log, tid, uuid)
	if foundMainImg {
		schema.put(mainImageField, mainImageUUID)
	}

	//embedded - image set(s), clip set(s), dynamic content(s) and custom code component(s) as embeds
	emContentUUIDs, foundEmbedded := extractEmbeddedContentByType(cc, u.log, acceptedTypes, tid, uuid)
	if foundEmbedded {
		schema.putAll(embeds, emContentUUIDs)
	}

	//promotional image from alternativeImages->promotionalImage as promotionalImage
	var foundPromImg bool
	altImg, found := cc[altImagesField].(map[string]interface{})
	if found {
		var promImg map[string]interface{}
		promImg, foundPromImg = altImg[promotionalImage].(map[string]interface{})
		if foundPromImg {
			if id, ok := promImg[id].(string); ok {
				u, err := extractUUIDFromString(id)
				if err != nil {
					localLog.WithError(err).Errorf("Cannot find promotional image: %v. Skipping expanding promotional image", err.Error())
					foundPromImg = false
				} else {
					schema.put(promotionalImage, u)
				}
			} else {
				localLog.Debug("Promotional image is missing the id field. Skipping expanding promotional image")
				foundPromImg = false
			}
		} else {
			localLog.Debug("Cannot find promotional image. Skipping expanding promotional image")
		}
	}

	if !foundMainImg && !foundEmbedded && !foundPromImg {
		localLog.Debugf("No main image or promotional image or embedded content to expand for supplied content %s", uuid)
		return nil
	}

	return schema
}

func (u *DefaultUnroller) resolveModelsForSetsMembers(s Schema, imgMap map[string]Content, tid string, uuid string) {
	mainImageUUID := s.get(mainImageField)
	u.resolveImageSet(mainImageUUID, imgMap, tid, uuid)
	// In embeds values we can have ImageSet, ClipSet, CustomCodeComponent or Dynamic Content.
	for _, embeddedImgSet := range s.getAll(embeds) {
		u.resolveImageSet(embeddedImgSet, imgMap, tid, uuid)
	}
}

func (u *DefaultUnroller) resolveImageSet(imageSetUUID string, imgMap map[string]Content, tid string, uuid string) {
	embedsFoundMember, found := resolveContent(imageSetUUID, imgMap)
	if !found {
		u.log.WithUUID(uuid).WithTransactionID(tid).Debugf("Cannot match to any found content UUID: %v", imageSetUUID)
		imgMap[imageSetUUID] = Content{id: createID(u.apiHost, "content", imageSetUUID)}
		return
	}

	localLog := u.log.WithUUID(uuid).WithTransactionID(tid)

	// ImageSet and ClipSet have members, and only they are processed in the following code
	rawMembers, found := embedsFoundMember[membersField]
	if found {
		membList, ok := rawMembers.([]interface{})
		if !ok {
			// u.log.Debugf("members field is present, but value is not a valid JSON in %v", uuid)
			return
		}

		expMembers := []Content{}
		for _, m := range membList {
			mData := fromMap(m.(map[string]interface{}))
			mID := mData[id].(string)
			mUUID, err := extractUUIDFromString(mID)
			if err != nil {
				localLog.WithError(err).Errorf("Error while extracting UUID from %s: %v", mID, err.Error())
				continue
			}
			mContent, found := resolveContent(mUUID, imgMap)
			if !found {
				expMembers = append(expMembers, mData)
				continue
			}
			if _, isPoster := mContent["poster"]; isPoster {
				resolvedPoster, err := u.resolvePoster(mContent["poster"], tid, uuid)
				if err != nil {
					localLog.WithError(err).Errorf("Error while getting expanded content for uuid: %s: %v", uuid, err.Error())
				} else {
					mContent["poster"] = resolvedPoster
				}
			}
			mData.merge(mContent)
			expMembers = append(expMembers, mData)
		}
		embedsFoundMember[membersField] = expMembers
	}
}

func (u *DefaultUnroller) resolvePoster(poster interface{}, tid, uuid string) (Content, error) {
	posterData, found := poster.(map[string]interface{})
	if !found {
		return Content{}, errors.New("problem in poster field")
	}
	papiurl := posterData[apiURLField].(string)
	pUUID, err := extractUUIDFromString(papiurl)
	if err != nil {
		return Content{}, err
	}
	posterContent, err := u.reader.Get([]string{pUUID}, tid)
	if err != nil {
		return Content{}, err
	}
	// Resolve poster is called by resolveImageSet, it is not OK to call back the same function
	// as it may cause an endless loop theoretically...
	u.resolveImageSet(pUUID, posterContent, tid, uuid)
	return posterContent[pUUID], nil
}

func (u *DefaultUnroller) resolveModelsForInnerBodyXML(
	s Schema,
	foundContent map[string]Content,
	acceptedTypes []string,
	tid string,
	uuid string) {
	for _, embedsContentUUID := range s.getAll(embeds) {
		embedsFoundContent, found := resolveContent(embedsContentUUID, foundContent)
		if !found {
			// UUID to ID transformation for not found content is already done in ImageSet transformation for all types.
			continue
		}

		localLog := u.log.WithUUID(uuid).WithTransactionID(tid)

		// Custom Code Component does not have members field, but it has bodyXML,
		// so we need to run one more check for inner ImageSet(s) and inner CCC(s)
		// similar to the extractEmbeddedContentByType in UniversalUnroller service.
		rawBody, foundBody := embedsFoundContent[bodyXMLField]
		if !foundBody {
			// localLog.Debug("Missing body. Skipping expanding embedded components and images")
			continue
		}

		bodyXML := rawBody.(string)
		emContentUUIDs, err := getEmbedded(u.log, bodyXML, acceptedTypes, tid, uuid)
		if err != nil {
			localLog.WithError(err).Errorf("failed to parse bodyXML for CCC content %s", err.Error())
			// Skip if the bodyXML of embeds component is corrupted, and it is not a valid XHTML.
			continue
		}

		if len(emContentUUIDs) == 0 {
			localLog.Debug("No embedded unrollable content inside the bodyXML")
			continue
		}

		// If these emContentUUIDs are already got, add the inner ImageSet and unroll its members,
		// and add the inner CustomCodeComponent to an inner embeds node.
		// Note if content type is CCC and it is already loaded this is a cycle!!!
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
				// append to embeds node already unrolled ImageSet with all members filled. The Same element if there are no members.
				innerEmbeds = append(innerEmbeds, expMembers)
			} else {
				newInnerContent = append(newInnerContent, innerUUID)
			}
		}
		// Read Inner Content For Unrolling via reader.go->Get (one more REST call)
		if len(newInnerContent) > 0 {
			tempContentMap, err := u.reader.Get(newInnerContent, tid)
			if err != nil {
				localLog.WithError(err).Infof("failed to read CustomCodeComponent inner content: %s", err)
				continue
			}
			// Append Inner Content to embeds content already found
			for _, newItemUUID := range newInnerContent {
				newInnerContentFound, found := resolveContent(newItemUUID, tempContentMap)
				if found {
					foundContent[newItemUUID] = newInnerContentFound
					// Expand members field of the new inner content, which was just loaded before append to embeds node
					expMembers, err := u.unrollMembersForImageSetInCCC(newInnerContentFound, tempContentMap, tid)
					if err != nil {
						localLog.Infof("failed to fill inner content members field. Check for ImageSet %s in CCC, which has not published images", newInnerContentFound)
						// TODO what to do in this case - we have CCC, it has ImageSet, but the image members are not found!!!
						innerEmbeds = append(innerEmbeds, newInnerContentFound)
						continue
					}
					// append to embeds node already unrolled ImageSet with all members filled. The Same element if there are no members.
					innerEmbeds = append(innerEmbeds, expMembers)
				} else {
					// We have an UUID in bodyXML, and it is not present in known resources, and
					// it is not found in a separate get call to downstream service. We cannot unroll it.
					// TODO shall we add only ID for it.
					localLog.Info("failed to find inner for CustomCodeComponent content for unroll")
				}
			}
		}
		// Add embeds element to CustomCodeComponent similar to Article, so the internal content is unrolled as well
		embedsFoundContent[embeds] = innerEmbeds
	}
}

func (u *DefaultUnroller) unrollMembersForImageSetInCCC(innerContent Content, loadedContent map[string]Content, tid string) (Content, error) {
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

func validateDefaultContent(content Content) bool {
	_, hasMainImage := content[mainImageField]
	_, hasBody := content[bodyXMLField]
	_, hasAltImg := content[altImagesField].(map[string]interface{})

	return hasMainImage || hasBody || hasAltImg
}
