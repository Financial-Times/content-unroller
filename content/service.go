package content

import (
	"github.com/pkg/errors"
)

const (
	ImageSetType       = "http://www.ft.com/ontology/content/ImageSet"
	DynamicContentType = "http://www.ft.com/ontology/content/DynamicContent"
	mainImage          = "mainImage"
	id                 = "id"
	embeds             = "embeds"
	altImages          = "alternativeImages"
	leadImages         = "leadImages"
	members            = "members"
	bodyXML            = "bodyXML"
	promotionalImage   = "promotionalImage"
	image              = "image"
)

type Unroller interface {
	UnrollContent(UnrollEvent) UnrollResult
	UnrollContentPreview(UnrollEvent) UnrollResult
	UnrollInternalContent(UnrollEvent) UnrollResult
	UnrollInternalContentPreview(UnrollEvent) UnrollResult
}

type ContentUnroller struct {
	reader  Reader
	apiHost string
}

type Content map[string]interface{}

type ContentSchema map[string][]string

func NewContentUnroller(r Reader, apiHost string) *ContentUnroller {
	return &ContentUnroller{
		reader:  r,
		apiHost: apiHost,
	}
}

func (u *ContentUnroller) UnrollContent(req UnrollEvent) UnrollResult {
	//make a copy of the content
	cc := req.c.clone()

	schema := u.createContentSchema(cc, []string{ImageSetType, DynamicContentType}, req.tid, req.uuid)
	if schema != nil {
		contentMap, err := u.reader.Get(schema.toArray(), req.tid)
		if err != nil {
			return UnrollResult{req.c, errors.Wrapf(err, "Error while getting expanded images for uuid:%v", req.uuid)}
		}
		u.resolveModelsForSetsMembers(schema, contentMap, req.tid, req.tid)

		mainImageUUID := schema.get(mainImage)
		if mainImageUUID != "" {
			cc[mainImage] = contentMap[mainImageUUID]
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
				cc[altImages].(map[string]interface{})[promotionalImage] = pi
			}
		}
	}

	return UnrollResult{cc, nil}
}

func (u *ContentUnroller) UnrollContentPreview(req UnrollEvent) UnrollResult {
	//make a copy of the content
	cc := req.c.clone()
	unrolledEmbedded := []Content{}

	schema := u.createContentSchema(cc, []string{ImageSetType}, req.tid, req.uuid)
	if schema != nil {
		contentMap, err := u.reader.Get(schema.toArray(), req.tid)
		if err != nil {
			return UnrollResult{req.c, errors.Wrapf(err, "Error while getting expanded images for uuid:%v", req.uuid)}
		}
		u.resolveModelsForSetsMembers(schema, contentMap, req.tid, req.tid)

		mainImageUUID := schema.get(mainImage)
		if mainImageUUID != "" {
			cc[mainImage] = contentMap[mainImageUUID]
		}

		// this is only for embedded images
		embeddedContentUUIDs := schema.getAll(embeds)
		if len(embeddedContentUUIDs) > 0 {
			for _, emb := range embeddedContentUUIDs {
				unrolledEmbedded = append(unrolledEmbedded, contentMap[emb])
			}
		}

		promImgUUID := schema.get(promotionalImage)
		if promImgUUID != "" {
			pi, found := contentMap[promImgUUID]
			if found {
				cc[altImages].(map[string]interface{})[promotionalImage] = pi
			}
		}
	}

	// unroll dynamic content from Native content source
	dynContents, foundDyn := u.unrollDynamicContent(cc, req.tid, req.uuid, u.reader.GetNative)
	if foundDyn {
		for _, dynC := range dynContents {
			unrolledEmbedded = append(unrolledEmbedded, dynC)
		}
	}

	if len(unrolledEmbedded) > 0 {
		cc[embeds] = unrolledEmbedded
	}

	return UnrollResult{cc, nil}
}

func (u *ContentUnroller) UnrollInternalContent(req UnrollEvent) UnrollResult {
	cc := req.c.clone()
	expLeadImages, foundImages := u.unrollLeadImages(cc, req.tid, req.uuid)
	if foundImages {
		cc[leadImages] = expLeadImages
	}

	dynContents, foundDyn := u.unrollDynamicContent(cc, req.tid, req.uuid, u.reader.GetInternal)
	if foundDyn {
		cc[embeds] = dynContents
	}

	return UnrollResult{cc, nil}
}

func (u *ContentUnroller) UnrollInternalContentPreview(req UnrollEvent) UnrollResult {
	cc := req.c.clone()
	expLeadImages, foundImages := u.unrollLeadImages(cc, req.tid, req.uuid)
	if foundImages {
		cc[leadImages] = expLeadImages
	}

	dynContents, foundDyn := u.unrollDynamicContent(cc, req.tid, req.uuid, u.reader.GetNative)
	if foundDyn {
		cc[embeds] = dynContents
	}

	return UnrollResult{cc, nil}
}

func (u *ContentUnroller) createContentSchema(cc Content, acceptedTypes []string, tid string, uuid string) ContentSchema {
	//mainImage
	schema := make(ContentSchema)
	mi, foundMainImg := cc[mainImage].(map[string]interface{})
	if foundMainImg {
		u, err := extractUUIDFromString(mi[id].(string))
		if err != nil {
			logger.Infof(tid, uuid, "Cannot find main image: %v. Skipping expanding main image", err.Error())
			foundMainImg = false
		} else {
			schema.put(mainImage, u)
		}
	} else {
		logger.Info(tid, uuid, "Cannot find main image. Skipping expanding main image")
	}

	//embedded - images and dynamic content
	emContentUUIDs, foundEmbedded := u.extractEmbeddedContentByType(cc, acceptedTypes, tid, uuid)
	if foundEmbedded {
		schema.putAll(embeds, emContentUUIDs)
	}

	//promotional image
	var foundPromImg bool
	altImg, found := cc[altImages].(map[string]interface{})
	if found {
		var promImg map[string]interface{}
		promImg, foundPromImg = altImg[promotionalImage].(map[string]interface{})
		if foundPromImg {
			if id, ok := promImg[id].(string); ok {
				u, err := extractUUIDFromString(id)
				if err != nil {
					logger.Infof(tid, uuid, "Cannot find promotional image: %v. Skipping expanding promotional image", err.Error())
					foundPromImg = false
				} else {
					schema.put(promotionalImage, u)
				}
			} else {
				logger.Info(tid, uuid, "Promotional image is missing the id field. Skipping expanding promotional image")
				foundPromImg = false
			}
		} else {
			logger.Info(tid, uuid, "Cannot find promotional image. Skipping expanding promotional image")
		}
	}

	if !foundMainImg && !foundEmbedded && !foundPromImg {
		logger.Infof(tid, uuid, "No main image or body images or promotional image to expand for supplied content %s", uuid)
		return nil
	}

	return schema
}

func (u *ContentUnroller) unrollLeadImages(cc Content, tid string, uuid string) ([]Content, bool) {
	images, foundLeadImages := cc[leadImages].([]interface{})
	if !foundLeadImages {
		logger.Info(tid, uuid, "No lead images to expand for supplied content")
		return nil, false
	}

	if len(images) == 0 {
		logger.Info(tid, uuid, "No lead images to expand for supplied content")
		return nil, false
	}
	schema := make(ContentSchema)
	for _, item := range images {
		li := item.(map[string]interface{})
		uuid, err := extractUUIDFromString(li[id].(string))
		if err != nil {
			logger.Infof(tid, uuid, "Error while getting UUID for %s: %v", li[id].(string), err.Error())
			continue
		}
		li[image] = uuid
		schema.put(leadImages, uuid)
	}

	imgMap, err := u.reader.Get(schema.toArray(), tid)
	if err != nil {
		logger.Errorf(tid, uuid, errors.Wrapf(err, "Error while getting content for expanded images uuid"))
		return nil, false
	}

	expLeadImages := []Content{}
	for _, li := range images {
		rawLi := li.(map[string]interface{})
		rawLiUUID := rawLi[image].(string)
		liContent := fromMap(rawLi)
		imageData, found := u.resolveContent(rawLiUUID, imgMap)
		if !found {
			logger.Infof(tid, uuid, "Missing image model %s. Returning only de id.", rawLiUUID)
			delete(liContent, image)
			expLeadImages = append(expLeadImages, liContent)
			continue
		}
		liContent[image] = imageData
		expLeadImages = append(expLeadImages, liContent)
	}

	cc[leadImages] = expLeadImages
	return expLeadImages, true
}

func (u *ContentUnroller) unrollDynamicContent(cc Content, tid string, uuid string, getContentFromSourceFn ReaderFunc) ([]Content, bool) {
	emContentUUIDs, foundEmbedded := u.extractEmbeddedContentByType(cc, []string{DynamicContentType}, tid, uuid)
	if !foundEmbedded {
		return nil, false
	}

	contentMap, err := getContentFromSourceFn(emContentUUIDs, tid)
	if err != nil {
		logger.Errorf(tid, uuid, errors.Wrapf(err, "Error while getting embedded dynamic content"))
		return nil, false
	}

	embedded := []Content{}
	for _, ec := range emContentUUIDs {
		embedded = append(embedded, contentMap[ec])
	}

	return embedded, true
}

func (u *ContentUnroller) resolveModelsForSetsMembers(b ContentSchema, imgMap map[string]Content, tid string, uuid string) {
	mainImageUUID := b.get(mainImage)
	u.resolveImageSet(mainImageUUID, imgMap, tid, uuid)
	for _, embeddedImgSet := range b.getAll(embeds) {
		u.resolveImageSet(embeddedImgSet, imgMap, tid, uuid)
	}
}

func (u *ContentUnroller) resolveImageSet(imageSetUUID string, imgMap map[string]Content, tid string, uuid string) {
	imageSet, found := u.resolveContent(imageSetUUID, imgMap)
	if !found {
		imgMap[imageSetUUID] = Content{id: createID(u.apiHost, "content", imageSetUUID)}
		return
	}

	rawMembers, found := imageSet[members]
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
				logger.Infof(tid, uuid, "Error while extracting UUID from %s: %v", mID, err.Error())
				continue
			}
			mContent, found := u.resolveContent(mUUID, imgMap)
			if !found {
				expMembers = append(expMembers, mData)
				continue
			}
			mData.merge(mContent)
			expMembers = append(expMembers, mData)
		}
		imageSet[members] = expMembers
	}

}

func (u *ContentUnroller) resolveContent(uuid string, imgMap map[string]Content) (Content, bool) {
	c, found := imgMap[uuid]
	if !found {
		return Content{}, false
	}
	return c, true
}

func (u *ContentUnroller) extractEmbeddedContentByType(cc Content, acceptedTypes []string, tid string, uuid string) ([]string, bool) {
	body, foundBody := cc[bodyXML]
	if !foundBody {
		logger.Info(tid, uuid, "Missing body. Skipping expanding embedded content and images.")
		return nil, false
	}

	bodyXML := body.(string)
	emContentUUIDs, err := getEmbedded(bodyXML, acceptedTypes, tid, uuid)
	if err != nil {
		logger.Errorf(tid, uuid, errors.Wrapf(err, "Cannot parse body for uuid=%s", uuid))
		return nil, false
	}

	if len(emContentUUIDs) == 0 {
		return nil, false
	}

	return emContentUUIDs, true
}

func (c Content) clone() Content {
	clone := make(Content)
	for k, v := range c {
		clone[k] = v
	}
	return clone
}

func (c Content) getMembersUUID() []string {
	uuids := []string{}
	members, found := c[members]
	if !found {
		return uuids
	}

	memList, ok := members.([]interface{})
	if !ok {
		return uuids
	}
	for _, m := range memList {
		mData := m.(map[string]interface{})
		url, found := mData[id].(string)
		if !found {
			continue
		}
		u, err := extractUUIDFromString(url)
		if err != nil {
			continue
		}
		uuids = append(uuids, u)
	}
	return uuids
}

func (c Content) merge(src Content) {
	for k, v := range src {
		c[k] = v
	}
}

func (u ContentSchema) put(key string, value string) {
	if key != mainImage && key != promotionalImage && key != leadImages {
		return
	}
	prev, found := u[key]
	if !found {
		u[key] = []string{value}
		return
	}
	act := append(prev, value)
	u[key] = act
}

func (u ContentSchema) get(key string) string {
	if _, found := u[key]; key != mainImage && key != promotionalImage || !found {
		return ""
	}
	return u[key][0]
}

func (u ContentSchema) putAll(key string, values []string) {
	if key != embeds && key != leadImages {
		return
	}
	prevValue, found := u[key]
	if !found {
		u[key] = values
		return
	}
	u[key] = append(prevValue, values...)
}

func (u ContentSchema) getAll(key string) []string {
	if key != embeds && key != leadImages {
		return []string{}
	}
	return u[key]
}

func (u ContentSchema) toArray() (UUIDs []string) {
	for _, v := range u {
		UUIDs = append(UUIDs, v...)
	}
	return UUIDs
}

func fromMap(src map[string]interface{}) Content {
	dest := Content{}
	for k, v := range src {
		dest[k] = v
	}
	return dest
}
