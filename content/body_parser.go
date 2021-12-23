package content

import (
	"strings"

	"golang.org/x/net/html"
)

func getEmbedded(body string, acceptedTypes []string, tid string, uuid string) ([]string, error) {
	embedsResult := []string{}
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return embedsResult, err
	}

	parse(doc, acceptedTypes, &embedsResult, tid, uuid)
	return embedsResult, nil
}

func parse(n *html.Node, acceptedTypes []string, embedsResult *[]string, tid string, uuid string) {
	if n.Data == "ft-content" {
		isEmbedded := false
		isTypeMatching := false
		var id string
		for _, a := range n.Attr {
			if a.Key == "data-embedded" && a.Val == "true" {
				isEmbedded = true
			} else if a.Key == "type" {
				isTypeMatching = isContentTypeMatching(a.Val, acceptedTypes)
			} else if a.Key == "url" {
				id = a.Val
			}
		}

		if isEmbedded && isTypeMatching {
			u, err := extractUUIDFromString(id)
			if err != nil {
				logger.Errorf(tid, uuid, "Cannot extract UUID: %v", err.Error())
			} else {
				*embedsResult = append(*embedsResult, u)
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		parse(c, acceptedTypes, embedsResult, tid, uuid)
	}
}

func isContentTypeMatching(contentType string, acceptedTypes []string) bool {
	for _, t := range acceptedTypes {
		if contentType == t {
			return true
		}
	}
	return false
}
