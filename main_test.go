package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Financial-Times/content-unroller/content"
	"github.com/Financial-Times/go-logger/v2"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const (
	contentStoreAppName = "content-source-app-name"
)

func TestContent_ShouldReturn200(t *testing.T) {
	contentStoreServiceMock := startContentServerMock("testdata/source-content-valid-response.json")
	srv := startUnrollerService(contentStoreServiceMock.URL)
	defer contentStoreServiceMock.Close()
	defer srv.Close()

	expected, err := os.ReadFile("testdata/content-valid-response.json")
	assert.NoError(t, err, "")

	body, err := os.ReadFile("testdata/content-valid-request.json")
	assert.NoError(t, err, "Cannot read file necessary for test case")
	resp, err := http.Post(srv.URL+"/content", "application/json", bytes.NewReader(body))
	assert.NoError(t, err, "Should not fail")
	defer resp.Body.Close()
	actualResponse, err := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NoError(t, err, "")
	assert.JSONEq(t, string(expected), string(actualResponse))
}

func TestContent_ShouldReturn400WhenInvalidJson(t *testing.T) {
	contentStoreServiceMock := startContentServerMock("testdata/source-content-valid-response.json")
	srv := startUnrollerService(contentStoreServiceMock.URL)
	defer contentStoreServiceMock.Close()
	defer srv.Close()

	body := `{"body":"invalid""body"}`
	resp, err := http.Post(srv.URL+"/content", "application/json", strings.NewReader(body))
	assert.NoError(t, err, "")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestContent_ShouldReturn400WhenInvalidContentRequest(t *testing.T) {
	contentStoreServiceMock := startContentServerMock("testdata/source-content-valid-response.json")
	srv := startUnrollerService(contentStoreServiceMock.URL)
	defer contentStoreServiceMock.Close()
	defer srv.Close()

	body := `{"id":"36037ab1-da3b-35bf-b5ee-4fc23723b635"}`
	resp, err := http.Post(srv.URL+"/content", "application/json", strings.NewReader(body))
	assert.NoError(t, err, "")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestInternalContent_ShouldReturn200(t *testing.T) {
	contentStoreServiceMock := startContentServerMock("testdata/internalcontent-source-valid-response.json")
	srv := startUnrollerService(contentStoreServiceMock.URL)
	defer contentStoreServiceMock.Close()
	defer srv.Close()

	expected, err := os.ReadFile("testdata/internalcontent-valid-response-no-lead-images.json")
	assert.NoError(t, err, "")

	body, err := os.ReadFile("testdata/internalcontent-valid-request.json")
	assert.NoError(t, err, "Cannot read file necessary for test case")
	resp, err := http.Post(srv.URL+"/internalcontent", "application/json", bytes.NewReader(body))
	assert.NoError(t, err, "Should not fail")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	actualResponse, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "")
	assert.JSONEq(t, string(expected), string(actualResponse))
}

func TestInternalContent_ShouldReturn400InvalidJson(t *testing.T) {
	internalContentStoreServiceMock := startContentServerMock("testdata/internalcontent-source-valid-response.json")
	srv := startUnrollerService(internalContentStoreServiceMock.URL)
	defer internalContentStoreServiceMock.Close()
	defer srv.Close()

	body := `{"body":"invalid""body"}`
	resp, err := http.Post(srv.URL+"/internalcontent", "application/json", strings.NewReader(body))
	assert.NoError(t, err, "")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestInternalContent_ShouldReturn400InvalidArticle(t *testing.T) {
	internalContentStoreServiceMock := startContentServerMock("testdata/internalcontent-source-valid-response.json")
	srv := startUnrollerService(internalContentStoreServiceMock.URL)
	defer internalContentStoreServiceMock.Close()
	defer srv.Close()

	body := `{"id":"36037ab1-da3b-35bf-b5ee-4fc23723b635"}`
	resp, err := http.Post(srv.URL+"/internalcontent", "application/json", strings.NewReader(body))
	assert.NoError(t, err, "")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestShouldBeHealthy(t *testing.T) {
	contentStoreServiceMock := startContentServerMock("testdata/source-content-valid-response.json")
	srv := startUnrollerService(contentStoreServiceMock.URL)
	defer contentStoreServiceMock.Close()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/__health")
	if err == nil {
		defer resp.Body.Close()
	}

	assert.NoError(t, err, "Cannot send request to health endpoint")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Response status should be 200")
}

func TestShouldBeGoodToGo(t *testing.T) {
	contentStoreServiceMock := startContentServerMock("testdata/source-content-valid-response.json")
	srv := startUnrollerService(contentStoreServiceMock.URL)
	defer contentStoreServiceMock.Close()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/__gtg")
	if err == nil {
		defer resp.Body.Close()
	}

	assert.NoError(t, err, "Cannot send request to gtg endpoint")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Response status should be 200")
}

func TestShouldNotBeGoodToGoWhenContentStoreIsNotHappy(t *testing.T) {
	contentStoreServiceMock := startUnhealthyContentServerMock()
	srv := startUnrollerService(contentStoreServiceMock.URL)
	defer contentStoreServiceMock.Close()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/__gtg")
	if err == nil {
		defer resp.Body.Close()
	}

	assert.NoError(t, err, "Cannot send request to gtg endpoint")
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode, "Response status should be 503")
}

func startContentServerMock(resource string) *httptest.Server {
	router := mux.NewRouter()
	router.Path("/__health").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(statusOkHandler)})
	router.Path("/__gtg").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(statusOkHandler)})
	router.Path("/").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(successfulContentServerMock(resource))})

	return httptest.NewServer(router)
}

func successfulContentServerMock(resource string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		file, err := os.Open(resource)
		if err != nil {
			return
		}
		defer file.Close()
		_, _ = io.Copy(w, file)
	}
}

func statusOkHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func startUnhealthyContentServerMock() *httptest.Server {
	router := mux.NewRouter()
	router.Path("/__health").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(statusServiceUnavailableHandler)})
	router.Path("/__gtg").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(statusServiceUnavailableHandler)})

	return httptest.NewServer(router)
}

func statusServiceUnavailableHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusServiceUnavailable)
}

func startUnrollerService(contentStoreURL string) *httptest.Server {
	sc := content.ServiceConfig{
		ContentStoreAppName:      contentStoreAppName,
		ContentStoreAppHealthURI: getServiceHealthURI(contentStoreURL),
		HTTPClient:               http.DefaultClient,
	}

	rc := content.ReaderConfig{
		ContentStoreAppName:         contentStoreAppName,
		ContentStoreHost:            contentStoreURL,
		ContentPathEndpoint:         "",
		InternalContentPathEndpoint: "",
	}

	reader := content.NewContentReader(rc, http.DefaultClient)
	testLogger := logger.NewUPPLogger("test-service", "Error")
	unroller := content.NewUniversalUnroller(reader, testLogger, contentStoreURL)
	handler := content.NewHandler(unroller, testLogger)

	h := setupServiceHandler(sc, *handler)
	return httptest.NewServer(h)
}
