package content

import (
	"fmt"
	"regexp"

	"errors"
)

const uuidRegex = "([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})"

func extractUUIDFromString(url string) (string, error) {
	re, err := regexp.Compile(uuidRegex)
	if err != nil {
		return "", errors.Join(err, fmt.Errorf("error during extracting UUID"))
	}

	values := re.FindStringSubmatch(url)
	if len(values) > 0 {
		return values[0], nil
	}
	return "", fmt.Errorf("cannot extract UUID from %s", url)
}

func createID(APIHost string, handlerPath string, uuid string) string {
	return "http://" + APIHost + "/" + handlerPath + "/" + uuid
}
