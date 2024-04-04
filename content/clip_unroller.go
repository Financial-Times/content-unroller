package content

type ClipUnroller struct {
	imageSetUnroller Unroller
	reader           Reader
	apiHost          string
}

func NewClipUnroller(imageSetUnroller Unroller, r Reader, apiHost string) *ClipUnroller {
	return &ClipUnroller{
		imageSetUnroller: imageSetUnroller,
		reader:           r,
		apiHost:          apiHost,
	}
}

func (u *ClipUnroller) Unroll(event UnrollEvent) (Content, error) {
	if !validateClip(event.c) {
		return nil, ValidationError
	}

	p, ok := event.c[posterField].(map[string]interface{})["apiUrl"].(string)
	if !ok {
		return nil, ConversionError
	}

	posterUUID, err := extractUUIDFromString(p)
	if err != nil {
		return nil, err
	}

	posterContent, err := u.reader.Get([]string{posterUUID}, event.tid)
	if err != nil {
		return nil, err
	}

	unrolledPoster, err := u.imageSetUnroller.Unroll(
		UnrollEvent{
			c:    posterContent[posterUUID],
			tid:  event.tid,
			uuid: posterUUID,
		},
	)
	if err != nil {
		return nil, err
	}

	returnContent := event.c.clone()
	returnContent[posterField] = unrolledPoster
	return returnContent, nil
}

func validateClip(clip Content) bool {
	return checkType(clip, ClipType)
}
