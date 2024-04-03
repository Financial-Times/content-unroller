package content

import (
	"errors"
)

var ConversionError = errors.New("failed to cast variable to expected type")

type ClipsetUnroller struct {
	clipUnroller Unroller
	reader       Reader
	apiHost      string
}

func NewClipsetUnroller(clipUnroller Unroller, r Reader, apiHost string) *ClipsetUnroller {
	return &ClipsetUnroller{
		clipUnroller: clipUnroller,
		reader:       r,
		apiHost:      apiHost,
	}
}

type member struct {
	UUID string `json:"id"`
}

func (u *ClipsetUnroller) Unroll(event UnrollEvent) (Content, error) {
	if !validateClipset(event.c) {
		return event.c, ValidationError
	}

	members, ok := event.c[membersField].([]member)
	if !ok {
		return nil, ConversionError
	}
	if len(members) == 0 {
		return event.c, nil
	}

	var clipUUIDs []string
	for _, m := range members {
		uuid, err := extractUUIDFromString(m.UUID)
		if err != nil {
			return nil, err
		}
		clipUUIDs = append(clipUUIDs, uuid)
	}

	clips, err := u.reader.Get(clipUUIDs, event.tid)
	if err != nil {
		return nil, err
	}

	var unrolledClips []Content
	for _, clipUUID := range clipUUIDs {
		unrolledClip, err := u.clipUnroller.Unroll(UnrollEvent{
			c:    clips[clipUUID],
			tid:  event.tid,
			uuid: clipUUID,
		})
		if err != nil {
			return nil, err
		}
		unrolledClips = append(unrolledClips, unrolledClip)
	}

	returnContent := event.c.clone()
	returnContent[membersField] = unrolledClips
	return returnContent, nil
}

func validateClipset(c Content) bool {
	_, hasMembers := c[membersField] //TODO: Check if clipset can have no members and will this field exists in this case

	return hasMembers && checkType(c, ClipSetType)
}
