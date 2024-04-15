package content

import "github.com/Financial-Times/go-logger/v2"

type DefaultInternalUnroller DefaultUnroller

func NewDefaultInternalUnroller(r Reader, log *logger.UPPLogger, apiHost string) *DefaultInternalUnroller {
	return (*DefaultInternalUnroller)(NewDefaultUnroller(r, log, apiHost))
}

func (u *DefaultInternalUnroller) Unroll(req UnrollEvent) (Content, error) {
	if !validateInternalDefaultContent(req.c) {
		return req.c, ErrValidating
	}

	cc := req.c.clone()
	expLeadImages, foundImages := unrollLeadImages(cc, u.reader, u.log, req.tid, req.uuid)
	if foundImages {
		cc[leadImages] = expLeadImages
	}

	dynContents, foundDyn := unrollDynamicContent(cc, u.log, req.tid, req.uuid, u.reader.GetInternal)
	if foundDyn {
		cc[embeds] = dynContents
	}

	return cc, nil
}

func validateInternalDefaultContent(content Content) bool {
	_, hasLeadImages := content[leadImages]
	_, hasBody := content[bodyXMLField]

	return hasLeadImages || hasBody
}
