package content

type InternalArticleUnroller ArticleUnroller

func NewInternalArticleUnroller(r Reader, apiHost string) *InternalArticleUnroller {
	return (*InternalArticleUnroller)(NewArticleUnroller(r, apiHost))
}

func (u *InternalArticleUnroller) Validate(article Content) bool {
	_, hasLeadImages := article[leadImages]
	_, hasBody := article[bodyXML]

	return hasLeadImages || hasBody
}

func (u *InternalArticleUnroller) Unroll(req UnrollEvent) UnrollResult {
	cc := req.c.clone()
	expLeadImages, foundImages := unrollLeadImages(cc, u.reader, req.tid, req.uuid)
	if foundImages {
		cc[leadImages] = expLeadImages
	}

	dynContents, foundDyn := unrollDynamicContent(cc, req.tid, req.uuid, u.reader.GetInternal)
	if foundDyn {
		cc[embeds] = dynContents
	}

	return UnrollResult{cc, nil}
}
