package content

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"errors"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/stretchr/testify/assert"
)

const (
	InvalidBodyRequest   = "{\"id\": \"d02886fc-58ff-11e8-9859-6668838a4c10\"}"
	invalidBodyMissingID = "{\"bodyXML\": \"sample body\"}"
)

type ContentUnrollerMock struct {
	mockUnrollContent func(UnrollEvent) (Content, error)
}

func (cu *ContentUnrollerMock) UnrollContent(event UnrollEvent) (Content, error) {
	return cu.mockUnrollContent(event)
}

func (cu *ContentUnrollerMock) UnrollInternalContent(event UnrollEvent) (Content, error) {
	return cu.mockUnrollContent(event)
}

func TestGetContentReturns200(t *testing.T) {
	cu := ContentUnrollerMock{
		mockUnrollContent: func(_ UnrollEvent) (Content, error) {
			var r Content
			fileBytes, err := os.ReadFile("testdata/content-valid-response.json")
			assert.NoError(t, err, "Cannot read resources test file")
			err = json.Unmarshal(fileBytes, &r)
			assert.NoError(t, err, "Cannot build json body")
			return r, nil
		},
	}

	h := Handler{
		Unroller: &cu,
		log:      logger.NewUPPLogger("test-service", "Error"),
	}
	body, err := os.ReadFile("testdata/content-valid-request.json")
	assert.NoError(t, err, "Cannot read test file")
	req, err := http.NewRequest(http.MethodPost, "/content", bytes.NewReader(body))
	assert.NoError(t, err, "Cannot create request necessary for test")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.GetContent)

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	expectedBody, err := os.ReadFile("testdata/content-valid-response.json")
	assert.NoError(t, err, "Cannot read test file")
	actualBody := rr.Body

	assert.JSONEq(t, string(expectedBody), string(actualBody.Bytes()))
}

func TestGetContent_UnrollEventError(t *testing.T) {
	h := Handler{
		Unroller: &UniversalUnroller{},
		log:      logger.NewUPPLogger("test-service", "Error"),
	}
	req, err := http.NewRequest(http.MethodPost, "/content", strings.NewReader("sample body"))
	assert.NoError(t, err, "Cannot create request necessary for test")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.GetContent)

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, string(rr.Body.Bytes()), "invalid character")
}

func TestGetContent_UnrollEventError_MissingID(t *testing.T) {
	h := Handler{
		Unroller: &UniversalUnroller{},
		log:      logger.NewUPPLogger("test-service", "Error"),
	}
	req, err := http.NewRequest(http.MethodPost, "/content", strings.NewReader(invalidBodyMissingID))
	assert.NoError(t, err, "Cannot create request necessary for test")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.GetContent)

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, string(rr.Body.Bytes()), "Missing or invalid id field")
}

func TestGetContent_ValidationError(t *testing.T) {
	h := Handler{
		Unroller: &UniversalUnroller{},
		log:      logger.NewUPPLogger("test-service", "Error"),
	}
	req, err := http.NewRequest(http.MethodPost, "/content", strings.NewReader(InvalidBodyRequest))
	assert.NoError(t, err, "Cannot create request necessary for test")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.GetContent)

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid content")
}

func TestGetContent_UnrollingError(t *testing.T) {
	cu := ContentUnrollerMock{
		mockUnrollContent: func(_ UnrollEvent) (Content, error) {
			return nil, errors.Join(ErrConnectingToAPI, fmt.Errorf("error while unrolling content"))
		},
	}

	h := Handler{
		Unroller: &cu,
		log:      logger.NewUPPLogger("test-service", "Error"),
	}
	body, err := os.ReadFile("testdata/content-valid-request.json")
	assert.NoError(t, err, "Cannot read test file")
	req, err := http.NewRequest(http.MethodPost, "/content", bytes.NewReader(body))
	assert.NoError(t, err, "Cannot create request necessary for test")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.GetContent)

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "error while unrolling content")
}

func TestGetInternalContentReturns200(t *testing.T) {
	cu := ContentUnrollerMock{
		mockUnrollContent: func(_ UnrollEvent) (Content, error) {
			var r Content
			fileBytes, err := os.ReadFile("testdata/internalcontent-valid-response.json")
			assert.NoError(t, err, "Cannot read test file")
			err = json.Unmarshal(fileBytes, &r)
			assert.NoError(t, err, "Cannot build json body")
			return r, nil
		},
	}

	h := Handler{
		Unroller: &cu,
		log:      logger.NewUPPLogger("test-service", "Error"),
	}
	body, err := os.ReadFile("testdata/internalcontent-valid-request.json")
	assert.NoError(t, err, "Cannot read test file")
	req, err := http.NewRequest(http.MethodPost, "/internalcontent", bytes.NewReader(body))
	assert.NoError(t, err, "Cannot create request necessary for test")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.GetInternalContent)

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	expectedBody, err := os.ReadFile("testdata/internalcontent-valid-response.json")
	assert.NoError(t, err, "Cannot read test file")
	assert.JSONEq(t, string(expectedBody), string(rr.Body.Bytes()))
}

func TestGetInternalContent_UnrollEventError(t *testing.T) {
	h := Handler{
		Unroller: &UniversalUnroller{},
		log:      logger.NewUPPLogger("test-service", "Error"),
	}
	req, err := http.NewRequest(http.MethodPost, "/internalcontent", strings.NewReader("sample body"))
	assert.NoError(t, err, "Cannot create request necessary for test")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.GetInternalContent)

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, string(rr.Body.Bytes()), "invalid character")
}

func TestGetInternalContent_UnrollEventError_MissingID(t *testing.T) {
	h := Handler{
		Unroller: &UniversalUnroller{},
		log:      logger.NewUPPLogger("test-service", "Error"),
	}
	req, err := http.NewRequest(http.MethodPost, "/internalcontent", strings.NewReader(invalidBodyMissingID))
	assert.NoError(t, err, "Cannot create request necessary for test")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.GetInternalContent)

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, string(rr.Body.Bytes()), "Missing or invalid id")
}

func TestGetInternalContent_ValidationError(t *testing.T) {
	h := Handler{
		Unroller: &UniversalUnroller{},
		log:      logger.NewUPPLogger("test-service", "Error"),
	}
	req, err := http.NewRequest(http.MethodPost, "/internalcontent", strings.NewReader(InvalidBodyRequest))
	assert.NoError(t, err, "Cannot create request necessary for test")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.GetInternalContent)

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid content")
}

func TestGetInternalContent_UnrollingError(t *testing.T) {
	cu := ContentUnrollerMock{
		mockUnrollContent: func(_ UnrollEvent) (Content, error) {
			return nil, errors.Join(ErrConnectingToAPI, fmt.Errorf("error while unrolling content"))
		},
	}

	h := Handler{
		Unroller: &cu,
		log:      logger.NewUPPLogger("test-service", "Error"),
	}
	body, err := os.ReadFile("testdata/internalcontent-valid-request.json")
	assert.NoError(t, err, "Cannot read test file")
	req, err := http.NewRequest(http.MethodPost, "/internalcontent", bytes.NewReader(body))
	assert.NoError(t, err, "Cannot create request necessary for test")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.GetInternalContent)

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "error while unrolling content")
}
