package content

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	unresolvedHostURL = "http://sampleAddress:8080/content"
	invalidHostURL    = "@$@"
)

var testData = []string{
	"639cd952-149f-11e7-2ea7-a07ecd9ac73f",
	"639cd952-149f-11e7-2ea7-a07ecd9ac73f",
	"71231d3a-13c7-11e7-2ea7-a07ecd9ac73f",
	"d02886fc-58ff-11e8-9859-6668838a4c10",
}

func successfulContentServerMock(t *testing.T, resource string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, err := os.Open(resource)
		if err != nil {
			assert.NoError(t, err, "File necessary for starting mock server not found.")
			return
		}
		defer file.Close()
		_, err = io.Copy(w, file)
		if err != nil {
			assert.NoError(t, err, "Cannot send file content to response writer for Content Mock Server.")
			return
		}
	}))
}

func errorContentServerMock(t *testing.T, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEqual(t, http.StatusOK, statusCode, fmt.Sprintf("Status code should not be %d", http.StatusOK))
		w.WriteHeader(statusCode)
	}))
}

func readerForTest(contentStoreHost string) *ContentReader {
	cfg := ReaderConfig{
		ContentStoreAppName: "content-source-app-name",
		ContentStoreHost:    contentStoreHost,
	}
	return NewContentReader(cfg, http.DefaultClient)
}

func TestGet(t *testing.T) {
	ts := successfulContentServerMock(t, "testdata/source-content-valid-response.json")
	defer ts.Close()

	cr := readerForTest(ts.URL)
	body, err := os.ReadFile("testdata/reader-content-valid-response.json")
	assert.NoError(t, err, "Cannot read file necessary for running test case.")

	var expected map[string]Content
	err = json.Unmarshal(body, &expected)
	assert.NoError(t, err, "Cannot read expected response for test case.")

	actual, err := cr.Get(testData, "tid_1")
	assert.NoError(t, err, "Error while getting content data")
	assert.Equal(t, expected, actual)
}

func TestGet_ContentSourceReturns500(t *testing.T) {
	ts := errorContentServerMock(t, http.StatusInternalServerError)
	defer ts.Close()

	cr := readerForTest(ts.URL)
	_, err := cr.Get(testData, "tid_1")
	assert.Error(t, err, "There should an error thrown")
}

func TestGet_ContentSourceReturns404(t *testing.T) {
	ts := errorContentServerMock(t, http.StatusNotFound)
	defer ts.Close()

	cr := readerForTest(ts.URL)
	_, err := cr.Get(testData, "tid_1")
	assert.Error(t, err, "There should an error thrown")
}

func TestGet_ContentSourceCannotBeResolved(t *testing.T) {
	cr := readerForTest(unresolvedHostURL)
	_, err := cr.Get(testData, "tid_1")
	assert.Error(t, err, "There should an error thrown")
}

func TestGet_ContentSourceHasInvalidURL(t *testing.T) {
	cr := readerForTest(invalidHostURL)
	_, err := cr.Get(testData, "tid_1")
	assert.Error(t, err, "There should an error thrown")
}

func TestGetInternal(t *testing.T) {
	ts := successfulContentServerMock(t, "testdata/internalcontent-source-valid-response.json")
	defer ts.Close()

	cr := readerForTest(ts.URL)
	body, err := os.ReadFile("testdata/reader-internalcontent-dynamic-valid-response.json")
	assert.NoError(t, err, "Cannot read file necessary for running test case.")

	var expected map[string]Content
	err = json.Unmarshal(body, &expected)
	assert.NoError(t, err, "Cannot read expected response for test case.")

	actual, err := cr.GetInternal(testData, "tid_1")
	assert.NoError(t, err, "Error while getting content data")
	assert.Equal(t, expected, actual)
}

func TestGetInternal_ContentSourceReturns500(t *testing.T) {
	ts := errorContentServerMock(t, http.StatusInternalServerError)
	defer ts.Close()

	cr := readerForTest(ts.URL)
	_, err := cr.GetInternal(testData, "tid_1")
	assert.Error(t, err, "There should an error thrown")
}

func TestGetInternal_ContentSourceReturns404(t *testing.T) {
	ts := errorContentServerMock(t, http.StatusNotFound)
	defer ts.Close()

	cr := readerForTest(ts.URL)
	_, err := cr.GetInternal(testData, "tid_1")
	assert.Error(t, err, "There should an error thrown")
}

func TestGetInternal_ContentSourceCannotBeResolved(t *testing.T) {
	cr := readerForTest(unresolvedHostURL)
	_, err := cr.GetInternal(testData, "tid_1")
	assert.Error(t, err, "There should an error thrown")
}

func TestGetInternal_ContentSourceHasInvalidURL(t *testing.T) {
	cr := readerForTest(invalidHostURL)
	_, err := cr.GetInternal(testData, "tid_1")
	assert.Error(t, err, "There should an error thrown")
}
