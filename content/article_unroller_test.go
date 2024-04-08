package content

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/stretchr/testify/assert"
)

func TestUnrollContent(t *testing.T) {
	cu := ArticleUnroller{
		reader: &ReaderMock{
			mockGet: func(c []string, tid string) (map[string]Content, error) {
				b, err := os.ReadFile("testdata/reader-content-valid-response.json")
				assert.NoError(t, err, "Cannot open file necessary for test case")
				var res map[string]Content
				err = json.Unmarshal(b, &res)
				assert.NoError(t, err, "Cannot return valid response")
				return res, nil
			},
		},
		log:     logger.NewUPPLogger("test-service", "Error"),
		apiHost: "test.api.ft.com",
	}

	expected, err := os.ReadFile("testdata/content-valid-response.json")
	assert.NoError(t, err, "Cannot read necessary test file")

	var c Content
	fileBytes, err := os.ReadFile("testdata/content-valid-request.json")
	assert.NoError(t, err, "Cannot read necessary test file")
	err = json.Unmarshal(fileBytes, &c)
	assert.NoError(t, err, "Cannot build json body")
	req := UnrollEvent{c, "tid_sample", "sample_uuid"}
	actual, actualErr := cu.Unroll(req)
	assert.NoError(t, actualErr, "Should not get an error when expanding images")

	actualJSON, err := json.Marshal(actual)
	assert.JSONEq(t, string(expected), string(actualJSON))
}

func TestUnrollContent_NilSchema(t *testing.T) {
	cu := ArticleUnroller{reader: nil}
	var c Content
	err := json.Unmarshal([]byte(InvalidBodyRequest), &c)
	assert.NoError(t, err, "Cannot build json body")

	req := UnrollEvent{c, "tid_sample", "sample_uuid"}
	actual, _ := cu.Unroll(req)
	actualJSON, err := json.Marshal(actual)

	assert.JSONEq(t, InvalidBodyRequest, string(actualJSON))
}

func TestUnrollContent_ErrorExpandingFromContentStore(t *testing.T) {
	cu := ArticleUnroller{
		reader: &ReaderMock{
			mockGet: func(c []string, tid string) (map[string]Content, error) {
				return nil, errors.New("Cannot expand content from content store")
			},
		},
		log:     logger.NewUPPLogger("test-service", "Error"),
		apiHost: "test.api.ft.com",
	}

	var c Content
	fileBytes, err := os.ReadFile("testdata/content-valid-request.json")
	assert.NoError(t, err, "Cannot read necessary test file")
	err = json.Unmarshal(fileBytes, &c)
	assert.NoError(t, err, "Cannot build json body")
	req := UnrollEvent{c, "tid_sample", "sample_uuid"}
	actual, actualErr := cu.Unroll(req)

	actualJSON, err := json.Marshal(actual)
	assert.JSONEq(t, string(fileBytes), string(actualJSON))
	assert.Error(t, actualErr, "Expected to return error when cannot read from content store")
}

func TestUnrollContent_SkipPromotionalImageWhenIdIsMissing(t *testing.T) {
	expectedAltImages := map[string]interface{}{
		"promotionalImage": map[string]interface{}{
			"": "http://api.ft.com/content/4723cb4e-027c-11e7-ace0-1ce02ef0def9",
		},
	}

	cu := ArticleUnroller{
		reader: &ReaderMock{
			mockGet: func(c []string, tid string) (map[string]Content, error) {
				b, err := os.ReadFile("testdata/reader-content-valid-response.json")
				assert.NoError(t, err, "Cannot open file necessary for test case")
				var res map[string]Content
				err = json.Unmarshal(b, &res)
				assert.NoError(t, err, "Cannot return valid response")
				return res, nil
			},
		},
		log:     logger.NewUPPLogger("test-service", "Error"),
		apiHost: "test.api.ft.com",
	}

	var c Content
	fileBytes, err := os.ReadFile("testdata/invalid-article-missing-promotionalImage-id.json")
	assert.NoError(t, err, "Cannot read necessary test file")
	err = json.Unmarshal(fileBytes, &c)
	assert.NoError(t, err, "Cannot build json body")
	req := UnrollEvent{c, "tid_sample", "sample_uuid"}
	actual, actualErr := cu.Unroll(req)

	assert.NoError(t, actualErr, "Should not get an error when expanding images")
	assert.Equal(t, expectedAltImages, actual[altImagesField])
}

func TestUnrollContent_SkipPromotionalImageWhenUUIDIsInvalid(t *testing.T) {
	expectedAltImages := map[string]interface{}{
		"promotionalImage": map[string]interface{}{
			"id": "http://api.ft.com/content/not-uuid",
		},
	}

	cu := ArticleUnroller{
		reader: &ReaderMock{
			mockGet: func(c []string, tid string) (map[string]Content, error) {
				b, err := os.ReadFile("testdata/reader-content-valid-response.json")
				assert.NoError(t, err, "Cannot open file necessary for test case")
				var res map[string]Content
				err = json.Unmarshal(b, &res)
				assert.NoError(t, err, "Cannot return valid response")
				return res, nil
			},
		},
		log:     logger.NewUPPLogger("test-service", "Error"),
		apiHost: "test.api.ft.com",
	}

	var c Content
	fileBytes, err := os.ReadFile("testdata/invalid-article-invalid-promotionalImage-uuid.json")
	assert.NoError(t, err, "Cannot read necessary test file")
	err = json.Unmarshal(fileBytes, &c)
	assert.NoError(t, err, "Cannot build json body")
	req := UnrollEvent{c, "tid_sample", "sample_uuid"}
	actual, actualErr := cu.Unroll(req)

	assert.NoError(t, actualErr, "Should not get an error when expanding images")
	assert.Equal(t, expectedAltImages, actual[altImagesField])
}

func TestUnrollContent_EmbeddedContentSkippedWhenMissingBodyXML(t *testing.T) {
	cu := ArticleUnroller{
		reader: &ReaderMock{
			mockGet: func(c []string, tid string) (map[string]Content, error) {
				b, err := os.ReadFile("testdata/reader-content-valid-response-no-body.json")
				assert.NoError(t, err, "Cannot open file necessary for test case")
				var res map[string]Content
				err = json.Unmarshal(b, &res)
				assert.NoError(t, err, "Cannot return valid response")
				return res, nil
			},
		},
		log:     logger.NewUPPLogger("test-service", "Error"),
		apiHost: "test.api.ft.com",
	}

	var c Content
	fileBytes, err := os.ReadFile("testdata/content-valid-request.json")
	assert.NoError(t, err, "Cannot read test file")
	err = json.Unmarshal(fileBytes, &c)
	c[bodyXMLField] = "invalid body"

	req := UnrollEvent{c, "tid_sample", "sample_uuid"}
	res, resErr := cu.Unroll(req)
	assert.NoError(t, resErr, "Should not receive error when body cannot be parsed.")
	assert.Nil(t, res["embeds"], "Response should not contain embeds field")
}
