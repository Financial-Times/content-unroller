package content

import (
	"errors"
	"fmt"
)

var ValidationError = errors.New("Invalid content")

type ArticleUnroller struct {
	reader  Reader
	apiHost string
}

func NewArticleUnroller(r Reader, apiHost string) *ArticleUnroller {
	return &ArticleUnroller{
		reader:  r,
		apiHost: apiHost,
	}
}

func (u *ArticleUnroller) Unroll(req UnrollEvent) (Content, error) {
	if !validateArticle(req.c) {
		return req.c, ValidationError
	}

	cc := req.c.clone()

	schema := u.createContentSchema(cc, []string{ImageSetType, DynamicContentType, ClipSetType}, req.tid, req.uuid)
	if schema != nil { //TODO: Invert check
		contentMap, err := u.reader.Get(schema.toArray(), req.tid)
		if err != nil {
			return req.c, errors.Join(err, fmt.Errorf("error while getting expanded content for uuid: %v", req.uuid))
		}
		u.resolveModelsForSetsMembers(schema, contentMap, req.tid, req.tid)

		mainImageUUID := schema.get(mainImageField)
		if mainImageUUID != "" {
			cc[mainImageField] = contentMap[mainImageUUID]
		}

		embeddedContentUUIDs := schema.getAll(embeds)
		if len(embeddedContentUUIDs) > 0 {
			embedded := []Content{}
			for _, emb := range embeddedContentUUIDs {
				embedded = append(embedded, contentMap[emb])
			}
			cc[embeds] = embedded
		}

		promImgUUID := schema.get(promotionalImage)
		if promImgUUID != "" {
			pi, found := contentMap[promImgUUID]
			if found {
				cc[altImagesField].(map[string]interface{})[promotionalImage] = pi
			}
		}
	}

	return cc, nil
}

func (u *ArticleUnroller) createContentSchema(cc Content, acceptedTypes []string, tid string, uuid string) ContentSchema {
	//mainImageField
	schema := make(ContentSchema)
	mi, foundMainImg := cc[mainImageField].(map[string]interface{})
	if foundMainImg {
		u, err := extractUUIDFromString(mi[id].(string))
		if err != nil {
			logger.Errorf(tid, uuid, "Cannot find main image: %v. Skipping expanding main image", err.Error())
			foundMainImg = false
		} else {
			schema.put(mainImageField, u)
		}
	} else {
		logger.Debug(tid, uuid, "Cannot find main image. Skipping expanding main image")
	}

	//embedded - images and dynamic content
	emContentUUIDs, foundEmbedded := extractEmbeddedContentByType(cc, acceptedTypes, tid, uuid)
	if foundEmbedded {
		schema.putAll(embeds, emContentUUIDs)
	}

	//promotional image
	var foundPromImg bool
	altImg, found := cc[altImagesField].(map[string]interface{})
	if found {
		var promImg map[string]interface{}
		promImg, foundPromImg = altImg[promotionalImage].(map[string]interface{})
		if foundPromImg {
			if id, ok := promImg[id].(string); ok {
				u, err := extractUUIDFromString(id)
				if err != nil {
					logger.Errorf(tid, uuid, "Cannot find promotional image: %v. Skipping expanding promotional image", err.Error())
					foundPromImg = false
				} else {
					schema.put(promotionalImage, u)
				}
			} else {
				logger.Debug(tid, uuid, "Promotional image is missing the id field. Skipping expanding promotional image")
				foundPromImg = false
			}
		} else {
			logger.Debug(tid, uuid, "Cannot find promotional image. Skipping expanding promotional image")
		}
	}

	if !foundMainImg && !foundEmbedded && !foundPromImg {
		logger.Debugf(tid, uuid, "No main image or promotional image or embedded content to expand for supplied content %s", uuid)
		return nil
	}

	return schema
}

func (u *ArticleUnroller) resolveModelsForSetsMembers(b ContentSchema, imgMap map[string]Content, tid string, uuid string) {
	mainImageUUID := b.get(mainImageField)
	u.resolveImageSet(mainImageUUID, imgMap, tid, uuid)
	for _, embeddedImgSet := range b.getAll(embeds) {
		u.resolveImageSet(embeddedImgSet, imgMap, tid, uuid)
	}
}

func (u *ArticleUnroller) resolveImageSet(imageSetUUID string, imgMap map[string]Content, tid string, uuid string) {
	imageSet, found := resolveContent(imageSetUUID, imgMap)
	if !found {
		imgMap[imageSetUUID] = Content{id: createID(u.apiHost, "content", imageSetUUID)}
		return
	}

	rawMembers, found := imageSet[membersField]
	if found {
		membList, ok := rawMembers.([]interface{})
		if !ok {
			return
		}

		expMembers := []Content{}
		for _, m := range membList {
			mData := fromMap(m.(map[string]interface{}))
			mID := mData[id].(string)
			mUUID, err := extractUUIDFromString(mID)
			if err != nil {
				logger.Errorf(tid, uuid, "Error while extracting UUID from %s: %v", mID, err.Error())
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
					logger.Errorf(tid, uuid, "Error while getting expanded content for uuid: %s: %v", uuid, err.Error())
				} else {
					mContent["poster"] = resolvedPoster
				}
			}
			mData.merge(mContent)
			expMembers = append(expMembers, mData)
		}
		imageSet[membersField] = expMembers
	}
}

func (u *ArticleUnroller) resolvePoster(poster interface{}, tid, uuid string) (Content, error) {
	posterData, found := poster.(map[string]interface{})
	if !found {
		return Content{}, errors.New("Problem in poster field")
	}
	papiurl := posterData["apiUrl"].(string)
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

func validateArticle(article Content) bool {
	_, hasMainImage := article[mainImageField]
	_, hasBody := article[bodyXMLField]
	_, hasAltImg := article[altImagesField].(map[string]interface{})
	contentType, _ := article[typeField].(string) //TODO: Add tests with types not containing article

	return (hasMainImage || hasBody || hasAltImg) && contentType == ArticleType
}
