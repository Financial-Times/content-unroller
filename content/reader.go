package content

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	transactionidutils "github.com/Financial-Times/transactionid-utils-go"
	uuidutils "github.com/Financial-Times/uuid-utils-go"
)

const (
	userAgent      = "User-Agent"
	userAgentValue = "UPP_content-unroller"
)

var APIConnectivityError = errors.New("error connecting to API")

type Reader interface {
	Get([]string, string) (map[string]Content, error)
	GetInternal([]string, string) (map[string]Content, error)
}

type ReaderFunc func([]string, string) (map[string]Content, error)

type ReaderConfig struct {
	ContentStoreAppName         string
	ContentStoreHost            string
	ContentPathEndpoint         string
	InternalContentPathEndpoint string
}

type ContentReader struct {
	client *http.Client
	config ReaderConfig
}

func NewContentReader(rConfig ReaderConfig, client *http.Client) *ContentReader {
	return &ContentReader{
		client: client,
		config: rConfig,
	}
}

// Get reads content from content-public-read
func (cr *ContentReader) Get(uuids []string, tid string) (map[string]Content, error) {
	var cm = make(map[string]Content)
	requestURL := fmt.Sprintf("%s%s", cr.config.ContentStoreHost, cr.config.ContentPathEndpoint)

	contentBatch, err := cr.doGet(uuids, tid, requestURL, cr.config.ContentStoreAppName)
	if err != nil {
		return cm, err
	}

	var imgModelUUIDs []string
	for _, c := range contentBatch {
		cr.addItemToMap(c, cm)
		if _, foundMembers := c[membersField]; foundMembers {
			imgModelUUIDs = append(imgModelUUIDs, c.getMembersUUID()...)
		}
	}

	if len(imgModelUUIDs) == 0 {
		return cm, nil
	}

	imgModelsList, err := cr.doGet(imgModelUUIDs, tid, requestURL, cr.config.ContentStoreAppName)
	if err != nil {
		return cm, err
	}

	for _, i := range imgModelsList {
		cr.addItemToMap(i, cm)
	}

	return cm, nil
}

// GetInternal reads internal components from content-public-read
func (cr *ContentReader) GetInternal(uuids []string, tid string) (map[string]Content, error) {
	var cm = make(map[string]Content)
	requestURL := fmt.Sprintf("%s%s", cr.config.ContentStoreHost, cr.config.InternalContentPathEndpoint)

	internalContent, err := cr.doGet(uuids, tid, requestURL, cr.config.ContentStoreAppName)
	if err != nil {
		return cm, err
	}

	for _, c := range internalContent {
		cr.addItemToMap(c, cm)
	}

	return cm, nil
}

func (cr *ContentReader) doGet(uuids []string, tid string, reqURL string, appName string) ([]Content, error) {
	var cb []Content

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return cb, errors.Join(err, APIConnectivityError, fmt.Errorf("error creating request to %v", appName))
	}

	req.Header.Add(transactionidutils.TransactionIDHeader, tid)
	req.Header.Set(userAgent, userAgentValue)
	q := req.URL.Query()
	for _, uuid := range uuids {
		if err = uuidutils.ValidateUUID(uuid); err == nil {
			q.Add("uuid", uuid)
		}
	}
	req.URL.RawQuery = q.Encode()
	res, err := cr.client.Do(req)
	if err != nil {
		return cb, errors.Join(err, APIConnectivityError, fmt.Errorf("request to %v failed", appName))
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return cb, errors.Join(APIConnectivityError, fmt.Errorf("request to %v failed with status code %d", appName, res.StatusCode))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return cb, errors.Join(err, APIConnectivityError, fmt.Errorf("error reading response received from %v", appName))
	}

	err = json.Unmarshal(body, &cb)
	if err != nil {
		return cb, errors.Join(err, APIConnectivityError, fmt.Errorf("error unmarshalling response from %v", appName))
	}
	return cb, nil
}

func (cr *ContentReader) addItemToMap(c Content, cm map[string]Content) {
	id, ok := c[id].(string)
	if !ok {
		return
	}
	uuid, err := extractUUIDFromString(id)
	if err != nil {
		return
	}
	cm[uuid] = c
}
