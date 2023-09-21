package content

import (
	"errors"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEmbedded(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		acceptedTypes  []string
		expectedOutput []string
		expectedErr    error
	}{
		{
			name:          "body with embedded clips should return slice of those clips uuids",
			body:          loadBodyFromFile(t, "testdata/bodyXml.xml"),
			acceptedTypes: []string{ClipSetType},
			expectedOutput: []string{
				"f6074f3c-b331-4a89-963c-f72eaf3895ae",
				"f5e1294c-7a47-4107-a5c4-9b6bfeb9efed",
			},
		},
		{
			name:           "body with no embedded clips should return empty slice",
			body:           "<body><p>Sample body</p></body>",
			acceptedTypes:  []string{ClipSetType},
			expectedOutput: []string{},
		},

		{
			name:          "body with embedded images should return slice of those image uuids",
			body:          loadBodyFromFile(t, "testdata/bodyXml.xml"),
			acceptedTypes: []string{ImageSetType},
			expectedOutput: []string{
				"639cd952-149f-11e7-2ea7-a07ecd9ac73f",
				"71231d3a-13c7-11e7-2ea7-a07ecd9ac73f",
				"0261ea4a-1474-11e7-1e92-847abda1ac65",
			},
		},
		{
			name:          "body with scrolly images should return slice of those image uuids",
			body:          loadBodyFromFile(t, "testdata/scrollyBodyXml.xml"),
			acceptedTypes: []string{ImageSetType},
			expectedOutput: []string{
				"37449625-c70b-4c22-8409-d3facb840df4",
				"05d93a7b-4fac-46e3-9ffd-d9ce70b73083",
			},
		},
		{
			name:           "body with no embedded images should return empty slice",
			body:           "<body><p>Sample body</p></body>",
			acceptedTypes:  []string{ImageSetType},
			expectedOutput: []string{},
		},
		{
			name:           "malformed body should return empty slice",
			body:           "Sample body",
			acceptedTypes:  []string{ImageSetType},
			expectedOutput: []string{},
		},
		{
			name:           "empty body should return empty slice",
			body:           "",
			acceptedTypes:  []string{ImageSetType, DynamicContentType},
			expectedOutput: []string{},
		},
		{
			name:          "body with embedded dynamic content should return slice of those dynamic content uuids",
			body:          loadBodyFromFile(t, "testdata/bodyXml.xml"),
			acceptedTypes: []string{DynamicContentType},
			expectedOutput: []string{
				"d02886fc-58ff-11e8-9859-6668838a4c10",
			},
		},
		{
			name:           "body with no embedded dynamic content should return empty slice",
			body:           "<body><p>Sample body</p></body>",
			acceptedTypes:  []string{ImageSetType},
			expectedOutput: []string{},
		},
		{
			name:          "body with embedded images & dynamic content should return slice of images & dynamic content uuids",
			body:          loadBodyFromFile(t, "testdata/bodyXml.xml"),
			acceptedTypes: []string{ImageSetType, DynamicContentType},
			expectedOutput: []string{
				"639cd952-149f-11e7-2ea7-a07ecd9ac73f",
				"71231d3a-13c7-11e7-2ea7-a07ecd9ac73f",
				"0261ea4a-1474-11e7-1e92-847abda1ac65",
				"d02886fc-58ff-11e8-9859-6668838a4c10",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			emImagesUUIDs, err := getEmbedded(test.body, test.acceptedTypes, "", "")
			if test.expectedErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, test.expectedErr))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedOutput, emImagesUUIDs)
			}
		})
	}
}

func loadBodyFromFile(t *testing.T, filePath string) string {
	t.Helper()
	data, err := ioutil.ReadFile(filePath)
	assert.NoError(t, err, "Cannot read test file")
	return string(data)
}
