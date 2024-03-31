package content

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const (
	ID         = "http://www.ft.com/thing/22c0d426-1466-11e7-b0c1-37e417ee6c76"
	expectedId = "22c0d426-1466-11e7-b0c1-37e417ee6c76"
)

type ReaderMock struct {
	mockGet         func(c []string, tid string) (map[string]Content, error)
	mockGetInternal func(uuids []string, tid string) (map[string]Content, error)
}

func (rm *ReaderMock) Get(c []string, tid string) (map[string]Content, error) {
	return rm.mockGet(c, tid)
}

func (rm *ReaderMock) GetInternal(c []string, tid string) (map[string]Content, error) {
	return rm.mockGetInternal(c, tid)
}

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
	actual := cu.Unroll(req)
	assert.NoError(t, actual.err, "Should not get an error when expanding images")

	actualJSON, err := json.Marshal(actual.uc)
	assert.JSONEq(t, string(expected), string(actualJSON))
}

func TestUnrollContent_NilSchema(t *testing.T) {
	cu := ArticleUnroller{reader: nil}
	var c Content
	err := json.Unmarshal([]byte(InvalidBodyRequest), &c)
	assert.NoError(t, err, "Cannot build json body")

	req := UnrollEvent{c, "tid_sample", "sample_uuid"}
	actual := cu.Unroll(req)
	actualJSON, err := json.Marshal(actual.uc)

	assert.JSONEq(t, InvalidBodyRequest, string(actualJSON))
}

func TestUnrollContent_ErrorExpandingFromContentStore(t *testing.T) {
	cu := ArticleUnroller{
		reader: &ReaderMock{
			mockGet: func(c []string, tid string) (map[string]Content, error) {
				return nil, errors.New("Cannot expand content from content store")
			},
		},
		apiHost: "test.api.ft.com",
	}

	var c Content
	fileBytes, err := os.ReadFile("testdata/content-valid-request.json")
	assert.NoError(t, err, "Cannot read necessary test file")
	err = json.Unmarshal(fileBytes, &c)
	assert.NoError(t, err, "Cannot build json body")
	req := UnrollEvent{c, "tid_sample", "sample_uuid"}
	actual := cu.Unroll(req)

	actualJSON, err := json.Marshal(actual.uc)
	assert.JSONEq(t, string(fileBytes), string(actualJSON))
	assert.Error(t, actual.err, "Expected to return error when cannot read from content store")
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
		apiHost: "test.api.ft.com",
	}

	var c Content
	fileBytes, err := os.ReadFile("testdata/invalid-article-missing-promotionalImage-id.json")
	assert.NoError(t, err, "Cannot read necessary test file")
	err = json.Unmarshal(fileBytes, &c)
	assert.NoError(t, err, "Cannot build json body")
	req := UnrollEvent{c, "tid_sample", "sample_uuid"}
	actual := cu.Unroll(req)

	assert.NoError(t, actual.err, "Should not get an error when expanding images")
	assert.Equal(t, expectedAltImages, actual.uc[altImages])
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
		apiHost: "test.api.ft.com",
	}

	var c Content
	fileBytes, err := os.ReadFile("testdata/invalid-article-invalid-promotionalImage-uuid.json")
	assert.NoError(t, err, "Cannot read necessary test file")
	err = json.Unmarshal(fileBytes, &c)
	assert.NoError(t, err, "Cannot build json body")
	req := UnrollEvent{c, "tid_sample", "sample_uuid"}
	actual := cu.Unroll(req)

	assert.NoError(t, actual.err, "Should not get an error when expanding images")
	assert.Equal(t, expectedAltImages, actual.uc[altImages])
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
		apiHost: "test.api.ft.com",
	}

	var c Content
	fileBytes, err := os.ReadFile("testdata/content-valid-request.json")
	assert.NoError(t, err, "Cannot read test file")
	err = json.Unmarshal(fileBytes, &c)
	c[bodyXML] = "invalid body"

	req := UnrollEvent{c, "tid_sample", "sample_uuid"}
	res := cu.Unroll(req)
	assert.NoError(t, res.err, "Should not receive error when body cannot be parsed.")
	assert.Nil(t, res.uc["embeds"], "Response should not contain embeds field")
}

//func TestUnrollContent_ClipSet(t *testing.T) {
//	cu := ArticleUnroller{
//		reader: &ReaderMock{
//			mockGet: func(c []string, tid string) (map[string]Content, error) {
//				b, err := os.ReadFile("testdata/reader-content-clipset-valid-response.json")
//				assert.NoError(t, err, "Cannot open file necessary for test case")
//				var res map[string]Content
//				err = json.Unmarshal(b, &res)
//				assert.NoError(t, err, "Cannot return valid response")
//				return res, nil
//			},
//		},
//		apiHost: "test.api.ft.com",
//	}
//
//	expected, err := os.ReadFile("testdata/content-clipset-valid-response.json")
//	assert.NoError(t, err, "Cannot read necessary test file")
//
//	var c Content
//	fileBytes, err := os.ReadFile("testdata/content-clipset-valid-request.json")
//	assert.NoError(t, err, "Cannot read necessary test file")
//	err = json.Unmarshal(fileBytes, &c)
//	assert.NoError(t, err, "Cannot build json body")
//	req := UnrollEvent{c, "tid_sample", "sample_uuid"}
//	actual := cu.Unroll(req)
//	assert.NoError(t, actual.err, "Should not get an error when expanding images")
//
//	actualJSON, err := json.Marshal(actual.uc)
//	assert.JSONEq(t, string(expected), string(actualJSON))
//}

func TestUnrollInternalContent(t *testing.T) {
	cu := InternalArticleUnroller{
		reader: &ReaderMock{
			mockGet: func(c []string, tid string) (map[string]Content, error) {
				b, err := os.ReadFile("testdata/reader-internalcontent-valid-response.json")
				assert.NoError(t, err, "Cannot open file necessary for test case")
				var res map[string]Content
				err = json.Unmarshal(b, &res)
				assert.NoError(t, err, "Cannot return valid response")
				return res, nil
			},
			mockGetInternal: func(c []string, tid string) (map[string]Content, error) {
				b, err := os.ReadFile("testdata/reader-internalcontent-dynamic-valid-response.json")
				assert.NoError(t, err, "Cannot open file necessary for test case")
				var res map[string]Content
				err = json.Unmarshal(b, &res)
				assert.NoError(t, err, "Cannot return valid response")
				return res, nil
			},
		},
		apiHost: "test.api.ft.com",
	}

	var c Content
	fileBytes, err := os.ReadFile("testdata/internalcontent-valid-request.json")
	assert.NoError(t, err, "File necessary for building request body nod found")
	err = json.Unmarshal(fileBytes, &c)

	expected, err := os.ReadFile("testdata/internalcontent-valid-response.json")
	assert.NoError(t, err, "Cannot read necessary test file")

	req := UnrollEvent{c, "tid_sample", "sample_uuid"}
	actual := cu.Unroll(req)
	assert.NoError(t, actual.err, "Should not receive error for expanding internal content")

	actualJSON, err := json.Marshal(actual.uc)
	assert.JSONEq(t, string(actualJSON), string(expected))
}

func TestUnrollInternalContent_LeadImagesSkippedWhenReadingError(t *testing.T) {
	cu := InternalArticleUnroller{
		reader: &ReaderMock{
			mockGet: func(c []string, tid string) (map[string]Content, error) {
				return nil, errors.New("Error retrieving content")
			},
			mockGetInternal: func(c []string, tid string) (map[string]Content, error) {
				b, err := os.ReadFile("testdata/reader-internalcontent-dynamic-valid-response.json")
				assert.NoError(t, err, "Cannot open file necessary for test case")
				var res map[string]Content
				err = json.Unmarshal(b, &res)
				assert.NoError(t, err, "Cannot return valid response")
				return res, nil
			},
		},
		apiHost: "test.api.ft.com",
	}

	var c Content
	fileBytes, err := os.ReadFile("testdata/internalcontent-valid-request.json")
	assert.NoError(t, err, "File necessary for building request body nod found")
	err = json.Unmarshal(fileBytes, &c)

	expected, err := os.ReadFile("testdata/internalcontent-valid-response-no-lead-images.json")
	assert.NoError(t, err, "Cannot read necessary test file")

	req := UnrollEvent{c, "tid_sample", "sample_uuid"}
	actual := cu.Unroll(req)
	assert.NoError(t, actual.err, "Should not receive error for expanding internal content")

	actualJSON, err := json.Marshal(actual.uc)
	assert.JSONEq(t, string(actualJSON), string(expected))
}

func TestUnrollInternalContent_DynamicContentSkippedWhenReadingError(t *testing.T) {
	cu := InternalArticleUnroller{
		reader: &ReaderMock{
			mockGet: func(c []string, tid string) (map[string]Content, error) {
				b, err := os.ReadFile("testdata/reader-internalcontent-valid-response.json")
				assert.NoError(t, err, "Cannot open file necessary for test case")
				var res map[string]Content
				err = json.Unmarshal(b, &res)
				assert.NoError(t, err, "Cannot return valid response")
				return res, nil
			},
			mockGetInternal: func(c []string, tid string) (map[string]Content, error) {
				return nil, errors.New("Error retrieving content")
			},
		},
		apiHost: "test.api.ft.com",
	}

	var c Content
	fileBytes, err := os.ReadFile("testdata/internalcontent-valid-request.json")
	assert.NoError(t, err, "File necessary for building request body nod found")
	err = json.Unmarshal(fileBytes, &c)

	expected, err := os.ReadFile("testdata/internalcontent-valid-response-no-dynamic-content.json")
	assert.NoError(t, err, "Cannot read necessary test file")

	req := UnrollEvent{c, "tid_sample", "sample_uuid"}
	actual := cu.Unroll(req)
	assert.NoError(t, actual.err, "Should not receive error for expanding internal content")

	actualJSON, err := json.Marshal(actual.uc)
	assert.JSONEq(t, string(actualJSON), string(expected))
}

func TestExtractIDFromURL(t *testing.T) {
	actual, err := extractUUIDFromString(ID)
	assert.NoError(t, err, "Test should not return error")
	assert.Equal(t, expectedId, actual, "Response id should be equal")
}
