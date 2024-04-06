package content

import "github.com/Financial-Times/go-logger/v2"

type InternalArticleUnroller ArticleUnroller

func NewInternalArticleUnroller(r Reader, log *logger.UPPLogger, apiHost string) *InternalArticleUnroller {
	return (*InternalArticleUnroller)(NewArticleUnroller(r, log, apiHost))
}

func (u *InternalArticleUnroller) Unroll(req UnrollEvent) (Content, error) {
	if !validateInternalArticle(req.c) {
		return req.c, ValidationError
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

func validateInternalArticle(article Content) bool {
	_, hasLeadImages := article[leadImages]
	_, hasBody := article[bodyXMLField]
	//TODO: Add tests with types not containing article

	return (hasLeadImages || hasBody) && checkType(article, ArticleType)
}
