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
	u.resolveModelsForInnerBodyXML(schema.getAll(embeds), contentMap, acceptedTypes, req.tid, req.uuid)

	// Add unrolled mainImage
	mainImageUUID := schema.get(mainImageField)
	if mainImageUUID != "" {
		cc[mainImageField] = contentMap[mainImageUUID]
	}

	// Add all unrolled embeds
	embeddedContentUUIDs := schema.getAll(embeds)
	if len(embeddedContentUUIDs) > 0 {
		var embedded []Content
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
	// Check: check if apiUrl is not found (!ok) in poster
	pUUID, err := extractUUIDFromString(papiurl)
	if err != nil {
		return Content{}, err
	}
	posterContent, err := u.reader.Get([]string{pUUID}, tid)
	if err != nil {
		return Content{}, err
	}
	u.resolveImageSet(pUUID, posterContent, tid, uuid)
	return posterContent[pUUID], nil
}

func (u *DefaultUnroller) resolveModelsForInnerBodyXML(
	embedsElements []string,
	foundContent map[string]Content,
	acceptedTypes []string,
	tid string,
	uuid string) {
	// Process CustomCodeComponent(s), but we process also the inner ImageSet or inner CCC if present.
	localLog := u.log.WithUUID(uuid).WithTransactionID(tid)

	for _, embeddedComponentUUID := range embedsElements {
		embedFoundMember, found := resolveContent(embeddedComponentUUID, foundContent)
		if !found {
			localLog.Debugf("cannot match to any found content UUID: %v", embeddedComponentUUID)
			foundContent[embeddedComponentUUID] = Content{id: createID(u.apiHost, "content", embeddedComponentUUID)}
			continue
		}

		// Custom Code Component does not have members field, but it has bodyXML,
		// so we need to run one more check for inner members and inner CCC
		// similar to the extractEmbeddedContentByType in UniversalUnroller service.
		rawBody, foundBody := embedFoundMember[bodyXMLField]
		if !foundBody {
			// localLog.Debug("Missing body. Skipping expanding embedded CCC and ImageSet.")
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
		// add the inner ImageSet (with Images/Graphics) and inner CustomCodeComponent to an inner embeds node.
		innerEmbeds, innerEmbedsUUIDs, err := processContentForEmbeds(emContentUUIDs, foundContent, u.reader, u.log, u.apiHost, tid, uuid)
		if err != nil {
			localLog.Infof("failed to load content to unroll in any of: %v", emContentUUIDs)
			continue
		}

		// Add embeds element to CustomCodeComponent similar to Article, so the internal content is unrolled as well
		embedFoundMember[embeds] = innerEmbeds

		u.resolveModelsForInnerBodyXML(innerEmbedsUUIDs, foundContent, acceptedTypes, tid, uuid)
	}
}

func validateDefaultContent(content Content) bool {
	_, hasMainImage := content[mainImageField]
	_, hasBody := content[bodyXMLField]
	_, hasAltImg := content[altImagesField].(map[string]interface{})

	return hasMainImage || hasBody || hasAltImg
}
