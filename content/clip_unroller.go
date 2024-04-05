package content

func (u *UniversalUnroller) unrollClip(event UnrollEvent) (Content, error) {
	if !validateClip(event.c) {
		return nil, ValidationError
	}

	p, ok := event.c[posterField]
	if !ok {
		return event.c, nil
	}

	apiUrl, ok := p.(map[string]interface{})["apiUrl"].(string)
	if !ok {
		return nil, ConversionError
	}

	posterUUID, err := extractUUIDFromString(apiUrl)
	if err != nil {
		return nil, err
	}

	posterContent, err := u.reader.Get([]string{posterUUID}, event.tid)
	if err != nil {
		return nil, err
	}

	unrolledPoster, err := u.unrollImageSet(
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
