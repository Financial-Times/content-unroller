package content

func (u *UniversalUnroller) unrollImageSet(event UnrollEvent) (Content, error) {
	if !validateImageSet(event.c) {
		return nil, ErrValidating
	}

	members, ok := event.c[membersField].([]interface{})
	if !ok {
		return nil, ErrConverting
	}
	if len(members) == 0 {
		return event.c, nil
	}

	var imageUUIDs []string
	for _, m := range members {
		memberID, ok := m.(map[string]interface{})["id"].(string)
		if !ok {
			return nil, ErrConverting
		}
		uuid, err := extractUUIDFromString(memberID)
		if err != nil {
			return nil, err
		}
		imageUUIDs = append(imageUUIDs, uuid)
	}

	images, err := u.reader.Get(imageUUIDs, event.tid)
	if err != nil {
		return nil, err
	}

	var unrolledImages []Content
	for _, imageUUID := range imageUUIDs {
		unrolledImages = append(unrolledImages, images[imageUUID])
	}

	returnContent := event.c.clone()
	returnContent[membersField] = unrolledImages
	return returnContent, nil
}

func validateImageSet(c Content) bool {
	_, ok := c[membersField]

	return ok && checkType(c, ImageSetType)
}
