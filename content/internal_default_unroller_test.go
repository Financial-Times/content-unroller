package content

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/stretchr/testify/assert"
)

func TestUnrollInternalContent(t *testing.T) {
	cu := DefaultInternalUnroller{
		reader: &ReaderMock{
			mockGet: func(_ []string, _ string) (map[string]Content, error) {
				b, err := os.ReadFile("testdata/reader-internalcontent-valid-response.json")
				assert.NoError(t, err, "Cannot open file necessary for test case")
				var res map[string]Content
				err = json.Unmarshal(b, &res)
				assert.NoError(t, err, "Cannot return valid response")
				return res, nil
			},
			mockGetInternal: func(_ []string, _ string) (map[string]Content, error) {
				b, err := os.ReadFile("testdata/reader-internalcontent-dynamic-valid-response.json")
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
	fileBytes, err := os.ReadFile("testdata/internalcontent-valid-request.json")
	assert.NoError(t, err, "File necessary for building request body nod found")
	err = json.Unmarshal(fileBytes, &c)
	assert.NoError(t, err, "Expected to build json body")

	expected, err := os.ReadFile("testdata/internalcontent-valid-response.json")
	assert.NoError(t, err, "Cannot read necessary test file")

	req := UnrollEvent{c, "tid_sample", "sample_uuid"}
	actual, actualErr := cu.Unroll(req)
	assert.NoError(t, actualErr, "Should not receive error for expanding internal content")

	actualJSON, err := json.Marshal(actual)
	assert.NoError(t, err, "Expected to marshall correctly")
	assert.JSONEq(t, string(actualJSON), string(expected))
}

func TestUnrollInternalContent_LeadImagesSkippedWhenReadingError(t *testing.T) {
	cu := DefaultInternalUnroller{
		reader: &ReaderMock{
			mockGet: func(_ []string, _ string) (map[string]Content, error) {
				return nil, errors.New("Error retrieving content")
			},
			mockGetInternal: func(_ []string, _ string) (map[string]Content, error) {
				b, err := os.ReadFile("testdata/reader-internalcontent-dynamic-valid-response.json")
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
	fileBytes, err := os.ReadFile("testdata/internalcontent-valid-request.json")
	assert.NoError(t, err, "File necessary for building request body nod found")
	err = json.Unmarshal(fileBytes, &c)
	assert.NoError(t, err, "Cannot build json body")

	expected, err := os.ReadFile("testdata/internalcontent-valid-response-no-lead-images.json")
	assert.NoError(t, err, "Cannot read necessary test file")

	req := UnrollEvent{c, "tid_sample", "sample_uuid"}
	actual, actualErr := cu.Unroll(req)
	assert.NoError(t, actualErr, "Should not receive error for expanding internal content")

	actualJSON, err := json.Marshal(actual)
	assert.NoError(t, err, "Expected to marshall correctly")
	assert.JSONEq(t, string(actualJSON), string(expected))
}

func TestUnrollInternalContent_DynamicContentSkippedWhenReadingError(t *testing.T) {
	cu := DefaultInternalUnroller{
		reader: &ReaderMock{
			mockGet: func(_ []string, _ string) (map[string]Content, error) {
				b, err := os.ReadFile("testdata/reader-internalcontent-valid-response.json")
				assert.NoError(t, err, "Cannot open file necessary for test case")
				var res map[string]Content
				err = json.Unmarshal(b, &res)
				assert.NoError(t, err, "Cannot return valid response")
				return res, nil
			},
			mockGetInternal: func(_ []string, _ string) (map[string]Content, error) {
				return nil, errors.New("Error retrieving content")
			},
		},
		log:     logger.NewUPPLogger("test-service", "Error"),
		apiHost: "test.api.ft.com",
	}

	var c Content
	fileBytes, err := os.ReadFile("testdata/internalcontent-valid-request.json")
	assert.NoError(t, err, "File necessary for building request body nod found")
	err = json.Unmarshal(fileBytes, &c)
	assert.NoError(t, err, "Cannot build json body")

	expected, err := os.ReadFile("testdata/internalcontent-valid-response-no-dynamic-content.json")
	assert.NoError(t, err, "Cannot read necessary test file")

	req := UnrollEvent{c, "tid_sample", "sample_uuid"}
	actual, actualErr := cu.Unroll(req)
	assert.NoError(t, actualErr, "Should not receive error for expanding internal content")

	actualJSON, err := json.Marshal(actual)
	assert.NoError(t, err, "Expected to marshall correctly")
	assert.JSONEq(t, string(actualJSON), string(expected))
}
