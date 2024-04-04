package content

type ImageSetUnroller struct {
	reader  Reader
	apiHost string
}

func NewImageSetUnroller(r Reader, apiHost string) *ImageSetUnroller {
	return &ImageSetUnroller{
		reader:  r,
		apiHost: apiHost,
	}
}

func (u *ImageSetUnroller) Unroll(event UnrollEvent) (Content, error) {
	if !validateImageSet(event.c) {
		return nil, ValidationError
	}

	members, ok := event.c[membersField].([]interface{})
	if !ok {
		return nil, ConversionError
	}
	if len(members) == 0 {
		return event.c, nil
	}

	var imageUUIDs []string
	for _, m := range members {
		memberID, ok := m.(map[string]interface{})["id"].(string)
		if !ok {
			return nil, ConversionError
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
	_, hasMembers := c[membersField] //TODO: Check if imageset can have no members and will this field exists in this case

	return hasMembers && checkType(c, ImageSetType)
}
