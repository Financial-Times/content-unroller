package content

type ImageSetUnroller struct {
	imageUnroller Unroller
	reader        Reader
	apiHost       string
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

	members, found := event.c[membersField].([]member)
	if !found {
		return nil, ConversionError
	}
	if len(members) == 0 {
		return event.c, nil
	}

	var imageUUIDs []string
	for _, m := range members {
		uuid, err := extractUUIDFromString(m.UUID)
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
		unrolledClip, err := u.imageUnroller.Unroll(UnrollEvent{
			c:    images[imageUUID],
			tid:  event.tid,
			uuid: imageUUID,
		})
		if err != nil {
			return nil, err
		}
		unrolledImages = append(unrolledImages, unrolledClip)
	}

	returnContent := event.c.clone()
	returnContent[membersField] = unrolledImages
	return returnContent, nil
}

func validateImageSet(c Content) bool {
	_, hasMembers := c[membersField] //TODO: Check if imageset can have no members and will this field exists in this case

	return hasMembers && checkType(c, ImageSetType)
}
