package content

func (u *UniversalUnroller) unrollClip(event UnrollEvent) (Content, error) {
	if !validateClip(event.c) {
		return nil, ErrValidating
	}

	p, ok := event.c[posterField]
	if !ok {
		return event.c, nil
	}

	posterMap, ok := p.(map[string]interface{})
	if !ok {
		return nil, ErrConverting
	}
	apiURL, ok := posterMap[apiURLField].(string)
	if !ok {
		return nil, ErrConverting
	}

	posterUUID, err := extractUUIDFromString(apiURL)
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
