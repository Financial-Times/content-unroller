package content

import (
	"errors"
	"slices"

	"github.com/Financial-Times/go-logger/v2"
)

const (
	CustomCodeComponentType = "http://www.ft.com/ontology/content/CustomCodeComponent"
	ImageSetType            = "http://www.ft.com/ontology/content/ImageSet"
	DynamicContentType      = "http://www.ft.com/ontology/content/DynamicContent"
	ClipSetType             = "http://www.ft.com/ontology/content/ClipSet"
	ClipType                = "http://www.ft.com/ontology/content/Clip"
	ArticleType             = "http://www.ft.com/ontology/content/Article"
	mainImageField          = "mainImage"
	id                      = "id"
	formatField             = "format"
	embeds                  = "embeds"
	altImagesField          = "alternativeImages"
	leadImages              = "leadImages"
	membersField            = "members"
	posterField             = "poster"
	bodyXMLField            = "bodyXML"
	promotionalImage        = "promotionalImage"
	image                   = "image"
	typeField               = "type"
	typesField              = "types"
	apiURLField             = "apiUrl"
)

var (
	ErrConverting = errors.New("failed to cast variable to expected type")
	ErrValidating = errors.New("invalid content")
)

type UniversalUnroller struct {
	reader  Reader
	log     *logger.UPPLogger
	apiHost string
}

func NewUniversalUnroller(r Reader, log *logger.UPPLogger, apiHost string) *UniversalUnroller {
	return &UniversalUnroller{
		reader:  r,
		log:     log,
		apiHost: apiHost,
	}
}

func (u *UniversalUnroller) UnrollContent(event UnrollEvent) (Content, error) {
	defaultUnroller := NewDefaultUnroller(u.reader, u.log, u.apiHost)

	switch getEventType(event.c) {
	case ClipSetType:
		return u.unrollClipSet(event)
	case ClipType:
		return u.unrollClip(event)
	case ImageSetType:
		return u.unrollImageSet(event)
	case CustomCodeComponentType:
		return u.unrollCustomCodeComponent(event)
	default:
		return defaultUnroller.Unroll(event)
	}
}

func (u *UniversalUnroller) UnrollInternalContent(event UnrollEvent) (Content, error) {
	defaultInternalUnroller := NewDefaultInternalUnroller(u.reader, u.log, u.apiHost)

	switch getEventType(event.c) {
	default:
		return defaultInternalUnroller.Unroll(event)
	}
}

type Content map[string]interface{}

func (c Content) clone() Content {
	clone := make(Content)
	for k, v := range c {
		clone[k] = v
	}
	return clone
}

func (c Content) getMembersUUID() []string {
	uuids := []string{}
	members, found := c[membersField]
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

// Schema is a map containing UUIDs of related content and the name of the field they are used in.
type Schema map[string][]string

func (u Schema) put(key string, value string) {
	if key != mainImageField && key != promotionalImage && key != leadImages {
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

func (u Schema) get(key string) string {
	if _, found := u[key]; key != mainImageField && key != promotionalImage || !found {
		return ""
	}
	return u[key][0]
}

func (u Schema) putAll(key string, values []string) {
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

func (u Schema) getAll(key string) []string {
	if key != embeds && key != leadImages {
		return []string{}
	}
	return u[key]
}

func (u Schema) toArray() (UUIDs []string) {
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

func unrollLeadImages(cc Content, r Reader, log *logger.UPPLogger, tid string, uuid string) ([]Content, bool) {
	localLog := log.WithTransactionID(tid).WithUUID(uuid)

	images, foundLeadImages := cc[leadImages].([]interface{})
	if !foundLeadImages {
		localLog.Debug("No lead images to expand for supplied content")
		return nil, false
	}

	if len(images) == 0 {
		localLog.Debug("No lead images to expand for supplied content")
		return nil, false
	}
	schema := make(Schema)
	for _, item := range images {
		li := item.(map[string]interface{})
		uuid, err := extractUUIDFromString(li[id].(string))
		if err != nil {
			localLog.WithError(err).Errorf("Error while getting UUID for %s: %v", li[id].(string), err.Error())
			continue
		}
		li[image] = uuid
		schema.put(leadImages, uuid)
	}

	imgMap, err := r.Get(schema.toArray(), tid)
	if err != nil {
		localLog.WithError(err).Errorf("Error while getting content for expanded images %s", err.Error())

		// couldn't get the images, so we have to delete the additional uuid field (previously added)
		for _, li := range images {
			rawLi := li.(map[string]interface{})
			delete(rawLi, image)
		}

		return nil, false
	}

	var expLeadImages []Content
	for _, li := range images {
		rawLi := li.(map[string]interface{})
		rawLiUUID := rawLi[image].(string)
		liContent := fromMap(rawLi)
		imageData, found := resolveContent(rawLiUUID, imgMap)
		if !found {
			localLog.Debugf("Missing image model %s. Returning only the id.", rawLiUUID)
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

func unrollDynamicContent(cc Content, log *logger.UPPLogger, tid string, uuid string, getContentFromSourceFn ReaderFunc) ([]Content, bool) {
	emContentUUIDs, foundEmbedded := extractEmbeddedContentByType(cc, log, []string{DynamicContentType}, tid, uuid)
	if !foundEmbedded {
		return nil, false
	}

	contentMap, err := getContentFromSourceFn(emContentUUIDs, tid)
	if err != nil {
		log.WithError(err).WithTransactionID(tid).WithUUID(uuid).Errorf(tid, "Error while getting embedded dynamic content %s", err.Error())
		return nil, false
	}

	var embedded []Content
	for _, ec := range emContentUUIDs {
		embedded = append(embedded, contentMap[ec])
	}

	return embedded, true
}

func resolveContent(uuid string, imgMap map[string]Content) (Content, bool) {
	c, found := imgMap[uuid]
	if !found {
		return Content{}, false
	}
	return c, true
}

func extractEmbeddedContentByType(cc Content, log *logger.UPPLogger, acceptedTypes []string, tid string, uuid string) ([]string, bool) {
	localLog := log.WithTransactionID(tid).WithUUID(uuid)

	body, foundBody := cc[bodyXMLField]
	if !foundBody {
		localLog.Debug("Missing body. Skipping expanding embedded content and images.")
		return nil, false
	}

	bodyXML := body.(string)
	emContentUUIDs, err := getEmbedded(log, bodyXML, acceptedTypes, tid, uuid)
	if err != nil {
		localLog.WithError(err).Errorf("Cannot parse bodyXML for content %s", err.Error())
		return nil, false
	}

	if len(emContentUUIDs) == 0 {
		return nil, false
	}

	return emContentUUIDs, true
}

func extractMainImageContentByType(cc Content, log *logger.UPPLogger, tid string, uuid string) (string, bool) {
	localLog := log.WithTransactionID(tid).WithUUID(uuid)
	mi, foundMainImg := cc[mainImageField].(map[string]interface{})
	if foundMainImg {
		u, err := extractUUIDFromString(mi[id].(string))
		if err != nil {
			localLog.WithError(err).Errorf("Cannot find main image: %v. Skipping expanding main image", err.Error())
			foundMainImg = false
		} else {
			return u, foundMainImg
		}
	} else {
		localLog.Debug(tid, uuid, "Cannot find main image. Skipping expanding main image")
	}
	return "", foundMainImg
}

func checkType(content Content, wantedType string) bool {
	if contentTypes, ok := content[typesField].([]interface{}); ok {
		return slices.ContainsFunc(contentTypes, func(contentType interface{}) bool {
			return contentType == wantedType
		})
	}
	contentType, _ := content[typeField].(string)
	return contentType == wantedType
}

func getEventType(content Content) string {
	if contentTypes, ok := content[typesField].([]interface{}); ok {
		if len(contentTypes) > 0 {
			return contentTypes[0].(string)
		}
		return ""
	}
	if t, ok := content[typeField].(string); ok {
		return t
	}

	return ""
}
