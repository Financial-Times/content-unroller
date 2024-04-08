package content

func (u *UniversalUnroller) unrollClipSet(event UnrollEvent) (Content, error) {
	if !validateClipset(event.c) {
		return nil, ValidationError
	}

	members, ok := event.c[membersField].([]interface{})
	if !ok {
		return nil, ConversionError
	}
	if len(members) == 0 {
		return event.c, nil
	}

	var clipUUIDs []string
	for _, m := range members {
		memberID, ok := m.(map[string]interface{})["id"].(string)
		if !ok {
			return nil, ConversionError
		}
		uuid, err := extractUUIDFromString(memberID)
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
		unrolledClip, err := u.unrollClip(UnrollEvent{
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
	_, hasMembers := c[membersField]

	return hasMembers && checkType(c, ClipSetType)
}
