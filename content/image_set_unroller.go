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

	// Endpoint for multiple UUIDs skip values if not found, but it returns 200 OK
	// If none of the requested UUIDs is found it would return empty result: []
	images, err := u.reader.Get(imageUUIDs, event.tid)
	if err != nil {
		return nil, err
	}

	unrolledImages := []Content{}
	for _, imageUUID := range imageUUIDs {
		unrolledMember := images[imageUUID]
		// Check if the member was found, before add it into members, and replace with an empty if not found
		// Even if it was: http://www.ft.com/thing/{{uuid}} - we replace it with /content/{{uuid}} when not found
		if unrolledMember == nil {
			unrolledMember = Content{id: createID(u.apiHost, "content", imageUUID)}
		}
		unrolledImages = append(unrolledImages, unrolledMember)
	}

	returnContent := event.c.clone()
	returnContent[membersField] = unrolledImages
	return returnContent, nil
}

func validateImageSet(c Content) bool {
	_, ok := c[membersField]

	return ok && checkType(c, ImageSetType)
}
