package content

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/stretchr/testify/assert"
)

func Test_validateCustomCodeComponent(t *testing.T) {
	tests := []struct {
		name    string
		content Content
		want    bool
	}{
		{
			name:    "empty",
			content: Content{},
			want:    false,
		},
		{
			name: "content-without-bodyXML",
			content: Content{
				typeField: CustomCodeComponentType,
			},
			want: false,
		},
		{
			name: "content-with-correct-type-empty-body",
			content: Content{
				typeField:    CustomCodeComponentType,
				bodyXMLField: "",
			},
			want: true,
		},
		{
			name: "content-with-correct-type-empty-body",
			content: Content{
				typeField:    CustomCodeComponentType,
				bodyXMLField: "<ft-content type=\"http://www.ft.com/ontology/content/CustomCodeComponent\" url=\"http://api-t.ft.com/content/11111111-1111-1111-1111-11111111111\" data-embedded=\"true\"> </ft-content>",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, validateCustomCodeComponent(tt.content), "validateCustomCodeComponent(%v)", tt.content)
		})
	}
}

func Test_UnrollCustomCodeComponent_DummyCase(t *testing.T) {
	testLogger := logger.NewUPPLogger("test-service", "Error")
	defaultAPIHost := "test.api.ft.com"
	testTID := "testTID"
	testUUID := "testUUID"
	invalidCCC := Content{
		typeField: CustomCodeComponentType,
		// missing bodyXML makes it invalid
	}
	validEmptyCCC := Content{
		typeField:    CustomCodeComponentType,
		bodyXMLField: "",
	}

	type fields struct {
		reader  Reader
		apiHost string
	}
	tests := []struct {
		name           string
		unrollerFields fields
		event          UnrollEvent
		want           Content
		wantErr        assert.ErrorAssertionFunc
	}{
		{
			name: "invalid-custom-code-component-missing-body-xml",
			unrollerFields: fields{
				reader:  nil,
				apiHost: defaultAPIHost,
			},
			event: UnrollEvent{
				c:    invalidCCC,
				tid:  testTID,
				uuid: testUUID,
			},
			want:    invalidCCC,
			wantErr: assert.Error,
		},

		{
			name: "valid-ccc-with-empty-bodyXML",
			unrollerFields: fields{
				reader:  nil,
				apiHost: defaultAPIHost,
			},
			event: UnrollEvent{
				c:    validEmptyCCC,
				tid:  testTID,
				uuid: testUUID,
			},
			want:    validEmptyCCC,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUniversalUnroller(tt.unrollerFields.reader, testLogger, tt.unrollerFields.apiHost)
			got, err := u.UnrollContent(tt.event)
			if !tt.wantErr(t, err, fmt.Sprintf("unrollCustomCodeComponent(%v)", tt.event)) {
				return
			}
			assert.Equalf(t, tt.want, got, "unrollCustomCodeComponent(%v)", tt.event)
		})
	}
}

func Test_UnrollCustomCodeComponent(t *testing.T) {
	testLogger := logger.NewUPPLogger("test-service", "Error")
	defaultAPIHost := "www.ft.com"
	testTID := "testTID"
	cccWithImageSetUUID := "ff056e44-f02d-4295-918f-12f38432c75a"
	imageSetUUID := "9d114020-f8d5-11ee-a53f-95f056a14e8f"
	memberOneUUID := "ac9c728f-b2d0-4651-bc27-b30fce158af5"
	memberTwoUUID := "06d72d8d-36b7-4e0c-a7ca-41cd529191e8"
	innerCCCUUID := "c8de2401-ef1c-47a9-bcce-b61194b1e6da"

	type fields struct {
		reader  Reader
		apiHost string
	}
	tests := []struct {
		name           string
		unrollerFields fields
		event          UnrollEvent
		want           string
		wantErr        assert.ErrorAssertionFunc
	}{
		// Real Unroll of the CustomCodeComponent
		{
			name: "valid-ccc-with-ImageSet-UUID-for-unroll",
			unrollerFields: fields{
				reader: &ReaderMock{
					mockGet: func(emContentUUIDs []string, _ string) (map[string]Content, error) {
						// If the request is for the Image Set, return it
						if slices.Contains(emContentUUIDs, imageSetUUID) {
							return map[string]Content{
								imageSetUUID: getTestContentFromFile(t, "testdata/ccc-with-image-set-and-fallback-get-image-set-response.json"),
							}, nil
						}
						// Return the ImageSet members as array
						members := getTestContentSliceFromFile(t, "testdata/ccc-with-image-set-and-fallback-get-members-response.json")
						membersMap := make(map[string]Content)
						membersMap[memberOneUUID] = members[0]
						membersMap[memberTwoUUID] = members[1]
						return membersMap, nil
					},
				},
				apiHost: defaultAPIHost,
			},
			event: UnrollEvent{
				c:    getTestContentFromFile(t, "testdata/ccc-with-image-set-and-fallback-valid.json"),
				tid:  testTID + "_" + cccWithImageSetUUID,
				uuid: cccWithImageSetUUID,
			},
			want:    loadBodyFromFile(t, "testdata/ccc-with-image-set-and-fallback-valid-unrolled.json"),
			wantErr: assert.NoError,
		},
		{
			name: "valid-ccc-with-ImageSet-not-found-members",
			unrollerFields: fields{
				reader: &ReaderMock{
					mockGet: func(emContentUUIDs []string, _ string) (map[string]Content, error) {
						// If the request is for the Image Set, return it
						if slices.Contains(emContentUUIDs, imageSetUUID) {
							return map[string]Content{
								imageSetUUID: getTestContentFromFile(t, "testdata/ccc-with-image-set-and-fallback-get-image-set-response.json"),
							}, nil
						}
						// Return Not Found members as empty map []
						membersMap := make(map[string]Content)
						return membersMap, nil
					},
				},
				apiHost: defaultAPIHost,
			},
			event: UnrollEvent{
				c:    getTestContentFromFile(t, "testdata/ccc-with-image-set-and-fallback-valid.json"),
				tid:  testTID + "_" + cccWithImageSetUUID,
				uuid: cccWithImageSetUUID,
			},
			want:    loadBodyFromFile(t, "testdata/ccc-with-image-set-and-fallback-valid-without-members-unrolled.json"),
			wantErr: assert.NoError,
		},
		{
			name: "valid-ccc-with-inner-ccc",
			unrollerFields: fields{
				reader: &ReaderMock{
					mockGet: func(emContentUUIDs []string, _ string) (map[string]Content, error) {
						// Return the Inner CCC and ImageSet
						if slices.Contains(emContentUUIDs, innerCCCUUID) {
							members := getTestContentSliceFromFile(t, "testdata/ccc-with-inner-ccc-get-inner-ccc-response.json")
							membersMap := make(map[string]Content)
							membersMap[innerCCCUUID] = members[0]
							membersMap[imageSetUUID] = members[1]
							return membersMap, nil
						}
						// Return the ImageSet members as array
						members := getTestContentSliceFromFile(t, "testdata/ccc-with-image-set-and-fallback-get-members-response.json")
						membersMap := make(map[string]Content)
						membersMap[memberOneUUID] = members[0]
						membersMap[memberTwoUUID] = members[1]
						return membersMap, nil
					},
				},
				apiHost: defaultAPIHost,
			},
			event: UnrollEvent{
				c:    getTestContentFromFile(t, "testdata/ccc-with-inner-ccc-valid.json"),
				tid:  testTID + "_" + cccWithImageSetUUID,
				uuid: cccWithImageSetUUID,
			},
			want:    loadBodyFromFile(t, "testdata/ccc-with-inner-ccc-valid-unrolled.json"),
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUniversalUnroller(tt.unrollerFields.reader, testLogger, tt.unrollerFields.apiHost)
			got, err := u.UnrollContent(tt.event)
			if !tt.wantErr(t, err, fmt.Sprintf("unrollCustomCodeComponent(%v)", tt.event)) {
				return
			}
			gotString := convertToString(t, got)
			assert.JSONEqf(t, tt.want, gotString, "unrollCustomCodeComponent(%v)", tt.event)
		})
	}
}

func convertToString(t *testing.T, got Content) string {
	gotJSON, err := json.Marshal(got)
	if err != nil {
		t.Errorf("failed to marshal got to JSON: %v", err)
	}
	return string(gotJSON)
}

// parseJSON is a Helper func to parse the String to Content struct
func parseJSON(t *testing.T, jsonStr string) (*Content, error) {
	t.Helper()
	var content Content
	err := json.Unmarshal([]byte(jsonStr), &content)
	if err != nil {
		return nil, err
	}
	return &content, nil
}

// parseJSONSlice is a Helper func to parse the JSON string into a slice of Content structs
func parseJSONSlice(t *testing.T, jsonStr string) ([]Content, error) {
	t.Helper()
	var contents []Content
	err := json.Unmarshal([]byte(jsonStr), &contents)
	if err != nil {
		return nil, err
	}
	return contents, nil
}

// getCCCWithImageSetAndFallbackOnly is a Helper func to read custom code component for
// unrolling from file and give it to the test function to unroll it as source.
func getTestContentFromFile(t *testing.T, testFile string) Content {
	t.Helper()
	if testFile == "" {
		return Content{}
	}
	testContentString := loadBodyFromFile(t, testFile)
	content, err := parseJSON(t, testContentString)
	if err != nil || content == nil {
		assert.NoError(t, err, "Cannot read test file %s", testFile)
	}
	if content == nil {
		content = &Content{}
	}
	return *content
}

func getTestContentSliceFromFile(t *testing.T, testFile string) []Content {
	t.Helper()
	if testFile == "" {
		return []Content{}
	}
	testContentString := loadBodyFromFile(t, testFile)
	content, err := parseJSONSlice(t, testContentString)
	if err != nil || content == nil {
		assert.NoError(t, err, "Cannot read test file %s", testFile)
	}
	if content == nil {
		content = []Content{}
	}
	return content
}
