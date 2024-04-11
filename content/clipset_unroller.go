package content

import (
	"errors"
	"fmt"
)

func (u *UniversalUnroller) unrollClipSet(event UnrollEvent) (Content, error) {
	clipSets, err := u.reader.Get([]string{event.uuid}, event.tid)
	readClipSet, ok := clipSets[event.uuid]
	if !ok {
		return nil, errors.Join(ErrValidating, fmt.Errorf("did not find clipset"))
	}
	event.c = readClipSet

	if !validateClipset(event.c) {
		return nil, ErrValidating
	}

	members, ok := event.c[membersField].([]interface{})
	if !ok {
		return nil, ErrConverting
	}
	if len(members) == 0 {
		return event.c, nil
	}

	clipUUIDAndFormat := map[string]string{} //TODO: This solution should be optimised to avoid using a map. Maybe using a single for loop can fix this.
	var clipUUIDs []string
	for _, m := range members {
		memberMap := m.(map[string]interface{})
		memberID, ok := memberMap["id"].(string)
		if !ok {
			return nil, ErrConverting
		}
		uuid, err := extractUUIDFromString(memberID)
		if err != nil {
			return nil, err
		}
		clipUUIDs = append(clipUUIDs, uuid)
		clipUUIDAndFormat[uuid], ok = memberMap[formatField].(string)
		if !ok {
			return nil, errors.Join(ErrConverting, fmt.Errorf("missing format field for clip %s", uuid))
		}
	}

	clips, err := u.reader.Get(clipUUIDs, event.tid)
	if err != nil {
		return nil, err
	}

	var unrolledClips []Content
	for _, clipUUID := range clipUUIDs {
		unrolledClip, err := u.unrollClip(UnrollEvent{
			c:    clips[clipUUID],
			tid:  event.tid,
			uuid: clipUUID,
		})
		if err != nil {
			return nil, err
		}
		unrolledClip[formatField] = clipUUIDAndFormat[clipUUID] //TODO: Reformat this
		unrolledClips = append(unrolledClips, unrolledClip)
	}

	returnContent := event.c.clone()
	returnContent[membersField] = unrolledClips
	return returnContent, nil
}

func validateClipset(c Content) bool {
	_, hasMembers := c[membersField]

	return hasMembers && checkType(c, ClipSetType)
}
