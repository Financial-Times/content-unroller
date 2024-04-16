package content

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/Financial-Times/go-logger/v2"
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

func TestUnrollContent_ClipSet(t *testing.T) {
	defaultReader := &ReaderMock{
		mockGet: func(_ []string, _ string) (map[string]Content, error) {
			b, err := os.ReadFile("testdata/reader-content-clipset-valid-response.json")
			assert.NoError(t, err, "Cannot open file necessary for test case")
			var res map[string]Content
			err = json.Unmarshal(b, &res)
			assert.NoError(t, err, "Cannot return valid response")
			return res, nil
		},
	}
	defaultAPIHost := "test.api.ft.com"
	unroller := UniversalUnroller{
		reader:  defaultReader,
		apiHost: defaultAPIHost,
	}

	expected, err := os.ReadFile("testdata/content-clipset-valid-response.json")
	assert.NoError(t, err, "Cannot read necessary test file")

	var c Content
	fileBytes, err := os.ReadFile("testdata/content-clipset-valid-request.json")
	assert.NoError(t, err, "Cannot read necessary test file")
	err = json.Unmarshal(fileBytes, &c)
	assert.NoError(t, err, "Cannot build json body")
	req := UnrollEvent{c, "tid_sample", "sample_uuid"}
	actual, err := unroller.UnrollContent(req)
	assert.NoError(t, err, "Should not get an error when expanding clipset")

	actualJSON, err := json.Marshal(actual)
	assert.NoError(t, err, "Expected to marshall correctly")
	assert.JSONEq(t, string(expected), string(actualJSON))
}

func TestExtractIDFromURL(t *testing.T) {
	actual, err := extractUUIDFromString(ID)
	assert.NoError(t, err, "Test should not return error")
	assert.Equal(t, expectedId, actual, "Response id should be equal")
}

func Test_checkType(t *testing.T) {
	type args struct {
		content    Content
		wantedType string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "missing_type",
			args: args{
				content:    Content{},
				wantedType: ClipSetType,
			},
			want: false,
		},
		{
			name: "wrong_type",
			args: args{
				content: Content{
					typeField: "wrong",
				},
				wantedType: ClipSetType,
			},
			want: false,
		},
		{
			name: "correct_type",
			args: args{
				content: Content{
					typeField: ClipSetType,
				},
				wantedType: ClipSetType,
			},
			want: true,
		},
		{
			name: "wrong_type_in_array",
			args: args{
				content: Content{
					typesField: []interface{}{"wrong"},
				},
				wantedType: ClipSetType,
			},
			want: false,
		},
		{
			name: "correct_type_in_array",
			args: args{
				content: Content{
					typesField: []interface{}{ClipSetType},
				},
				wantedType: ClipSetType,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, checkType(tt.args.content, tt.args.wantedType), "checkType(%v, %v)", tt.args.content, tt.args.wantedType)
		})
	}
}

func Test_getEventType(t *testing.T) {
	tests := []struct {
		name    string
		content Content
		want    string
	}{
		{
			name:    "missing_type",
			content: Content{},
			want:    "",
		},
		{
			name: "correct_type",
			content: Content{
				typeField: ClipSetType,
			},
			want: ClipSetType,
		},
		{
			name: "correct_type_in_array",
			content: Content{
				typesField: []interface{}{ClipSetType},
			},
			want: ClipSetType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getEventType(tt.content), "getEventType(%v)", tt.content)
		})
	}
}

func TestUnrollContent_LiveBlogPackage(t *testing.T) {
	testReader := &ReaderMock{
		mockGet: func(_ []string, _ string) (map[string]Content, error) {
			b, err := os.ReadFile("testdata/reader-content-liveblogpackage-valid-response.json")
			assert.NoError(t, err, "Cannot open file necessary for test case")
			var res map[string]Content
			err = json.Unmarshal(b, &res)
			assert.NoError(t, err, "Cannot return valid response")
			return res, nil
		},
	}
	defaultAPIHost := "test.api.ft.com"
	unroller := UniversalUnroller{
		reader:  testReader,
		log:     logger.NewUPPLogger("test", "debug"),
		apiHost: defaultAPIHost,
	}

	expected, err := os.ReadFile("testdata/content-liveblogpackage-valid-response.json")
	assert.NoError(t, err, "Cannot read necessary test file")

	var c Content
	fileBytes, err := os.ReadFile("testdata/content-liveblogpackage-valid-request.json")
	assert.NoError(t, err, "Cannot read necessary test file")
	err = json.Unmarshal(fileBytes, &c)
	assert.NoError(t, err, "Cannot build json body")
	req := UnrollEvent{c, "tid_sample", "sample_uuid"}
	actual, err := unroller.UnrollContent(req)
	assert.NoError(t, err, "Should not get an error when expanding clipset")

	actualJSON, err := json.Marshal(actual)
	assert.NoError(t, err, "Expected to marshall correctly")
	assert.JSONEq(t, string(expected), string(actualJSON))
}

func TestUnrollContent_ContentPackage(t *testing.T) {
	testReader := &ReaderMock{
		mockGet: func(_ []string, _ string) (map[string]Content, error) {
			b, err := os.ReadFile("testdata/reader-internalcontent-contentpackage-valid-response.json")
			assert.NoError(t, err, "Cannot open file necessary for test case")
			var res map[string]Content
			err = json.Unmarshal(b, &res)
			assert.NoError(t, err, "Cannot return valid response")
			return res, nil
		},
	}
	defaultAPIHost := "test.api.ft.com"
	unroller := UniversalUnroller{
		reader:  testReader,
		log:     logger.NewUPPLogger("test", "debug"),
		apiHost: defaultAPIHost,
	}

	expected, err := os.ReadFile("testdata/internalcontent-contentpackage-valid-response.json")
	assert.NoError(t, err, "Cannot read necessary test file")

	var c Content
	fileBytes, err := os.ReadFile("testdata/internalcontent-contentpackage-valid-request.json")
	assert.NoError(t, err, "Cannot read necessary test file")
	err = json.Unmarshal(fileBytes, &c)
	assert.NoError(t, err, "Cannot build json body")
	req := UnrollEvent{c, "tid_sample", "sample_uuid"}
	actual, err := unroller.UnrollInternalContent(req)
	assert.NoError(t, err, "Should not get an error when expanding clipset")

	actualJSON, err := json.Marshal(actual)
	assert.NoError(t, err, "Expected to marshall correctly")
	assert.JSONEq(t, string(expected), string(actualJSON))
}
