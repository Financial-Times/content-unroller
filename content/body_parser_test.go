package content

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldReturnImages(t *testing.T) {
	var expectedOutput = []string{
		"639cd952-149f-11e7-2ea7-a07ecd9ac73f",
		"71231d3a-13c7-11e7-2ea7-a07ecd9ac73f",
		"0261ea4a-1474-11e7-1e92-847abda1ac65",
	}

	fileBytes, err := ioutil.ReadFile("../testdata/bodyXml.xml")
	if err != nil {
		assert.Fail(t, "Cannot read test file")
	}
	str := string(fileBytes)
	emImagesUUIDs, err := getEmbedded(str, []string{ImageSetType}, "", "")
	if err != nil {
		assert.Fail(t, err.Error())
	}

	assert.Equal(t, expectedOutput, emImagesUUIDs, "Response image ids should be equal to expected images")
}

func TestShouldReturnScrollyImages(t *testing.T) {
	var expectedOutput = []string{
		"37449625-c70b-4c22-8409-d3facb840df4",
		"05d93a7b-4fac-46e3-9ffd-d9ce70b73083",
	}

	fileBytes, err := ioutil.ReadFile("../testdata/scrollyBodyXml.xml")
	if err != nil {
		assert.Fail(t, "Cannot read test file")
	}
	str := string(fileBytes)
	emImagesUUIDs, err := getEmbedded(str, []string{ImageSetType}, "", "")
	if err != nil {
		assert.Fail(t, err.Error())
	}

	assert.Equal(t, expectedOutput, emImagesUUIDs, "Response image ids should be equal to expected images")
}

func TestBodyNoEmbeddedImagesReturnsEmptyList(t *testing.T) {
	emImagesUUIDs, err := getEmbedded("<body><p>Sample body</p></body>", []string{ImageSetType}, "", "")
	assert.NoError(t, err, "Body parsing should be successful")
	assert.Len(t, emImagesUUIDs, 0, "Response image ids should be equal to expected images")
}

func TestMalformedBodyReturnsEmptyList(t *testing.T) {
	emImagesUUIDs, err := getEmbedded("Sample body", []string{ImageSetType}, "", "")
	assert.NoError(t, err, "Body parsing should be successful")
	assert.Len(t, emImagesUUIDs, 0, "Response image ids should be equal to expected images")
}

func TestEmptyBodyReturnsEmptyList(t *testing.T) {
	emImagesUUIDs, _ := getEmbedded("", []string{ImageSetType, DynamicContentType}, "", "")
	assert.Equal(t, 0, len(emImagesUUIDs), "Response should return zero images")
}

func TestShouldReturnDynamicContent(t *testing.T) {
	var expectedOutput = []string{
		"d02886fc-58ff-11e8-9859-6668838a4c10",
	}

	fileBytes, err := ioutil.ReadFile("../testdata/bodyXml.xml")
	if err != nil {
		assert.Fail(t, "Cannot read test file")
	}
	str := string(fileBytes)
	emDynContentUUIDs, err := getEmbedded(str, []string{DynamicContentType}, "", "")
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.Equal(t, expectedOutput, emDynContentUUIDs, "Embedded dynamic content not extracted correctly from bodyXML")
}

func TestBodyNoEmbeddedDynamicContentReturnsEmptyList(t *testing.T) {
	emImagesUUIDs, err := getEmbedded("<body><p>Sample body</p></body>", []string{DynamicContentType}, "", "")
	assert.NoError(t, err, "Body parsing should be successful")
	assert.Len(t, emImagesUUIDs, 0, "Response image ids should be equal to expected images")
}

func TestShouldReturnImagesAndDynamicContent(t *testing.T) {
	var expectedOutput = []string{
		"639cd952-149f-11e7-2ea7-a07ecd9ac73f",
		"71231d3a-13c7-11e7-2ea7-a07ecd9ac73f",
		"0261ea4a-1474-11e7-1e92-847abda1ac65",
		"d02886fc-58ff-11e8-9859-6668838a4c10",
	}

	fileBytes, err := ioutil.ReadFile("../testdata/bodyXml.xml")
	if err != nil {
		assert.Fail(t, "Cannot read test file")
	}
	str := string(fileBytes)
	emImagesUUIDs, err := getEmbedded(str, []string{ImageSetType, DynamicContentType}, "", "")
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.Equal(t, expectedOutput, emImagesUUIDs, "Response image ids should be equal to expected images")
}
