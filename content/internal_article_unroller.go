package content

import (
	"slices"
)

type InternalArticleUnroller ArticleUnroller

func NewInternalArticleUnroller(r Reader, apiHost string) *InternalArticleUnroller {
	return &InternalArticleUnroller{
		reader:  r,
		apiHost: apiHost,
	}
}

func (u *InternalArticleUnroller) Unroll(req UnrollEvent) (Content, error) {
	if !validateInternalArticle(req.c) {
		return req.c, ValidationError
	}

	cc := req.c.clone()
	expLeadImages, foundImages := unrollLeadImages(cc, u.reader, req.tid, req.uuid)
	if foundImages {
		cc[leadImages] = expLeadImages
	}

	dynContents, foundDyn := unrollDynamicContent(cc, req.tid, req.uuid, u.reader.GetInternal)
	if foundDyn {
		cc[embeds] = dynContents
	}

	return cc, nil
}

func validateInternalArticle(article Content) bool {
	_, hasLeadImages := article[leadImages]
	_, hasBody := article[bodyXMLField]
	contentTypes, _ := article[typesField].([]interface{}) //TODO: Add tests with types not containing article

	return (hasLeadImages || hasBody) && slices.Contains(contentTypes, ArticleType)
}
