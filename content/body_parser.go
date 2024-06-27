package content

import (
	"slices"
	"strings"

	"github.com/Financial-Times/go-logger/v2"
	"golang.org/x/net/html"
)

func getEmbedded(log *logger.UPPLogger, body string, acceptedTypes []string, tid string, uuid string) ([]string, error) {
	embedsResult := []string{}
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return embedsResult, err
	}

	parse(doc, log, acceptedTypes, &embedsResult, tid, uuid)
	return embedsResult, nil
}

func parse(n *html.Node, log *logger.UPPLogger, acceptedTypes []string, embedsResult *[]string, tid string, uuid string) {
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
				log.WithError(err).Errorf(tid, uuid, "Cannot extract UUID: %v", err.Error())
			} else {
				*embedsResult = append(*embedsResult, u)
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		parse(c, log, acceptedTypes, embedsResult, tid, uuid)
	}
}

func isContentTypeMatching(contentType string, acceptedTypes []string) bool {
	return slices.Contains(acceptedTypes, contentType)
}
