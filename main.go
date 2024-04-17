package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/Financial-Times/content-unroller/content"
	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	cli "github.com/jawher/mow.cli"
)

const (
	AppCode = "content-unroller"
	AppName = "Content Unroller"
	AppDesc = "Content Unroller - unroll images and dynamic content for a given content"
)

func main() {
	app := cli.App(AppCode, AppDesc)
	port := app.String(cli.StringOpt{
		Name:   "port",
		Value:  "9090",
		Desc:   "application port",
		EnvVar: "PORT",
	})
	contentStoreApplicationName := app.String(cli.StringOpt{
		Name:   "contentSourceAppName",
		Value:  "content-public-read",
		Desc:   "Content read app",
		EnvVar: "CONTENT_STORE_APP_NAME",
	})
	contentStoreHost := app.String(cli.StringOpt{
		Name:   "contentStoreHost",
		Value:  "http://localhost:8080/__content-public-read",
		Desc:   "Content source hostname",
		EnvVar: "CONTENT_STORE_HOST",
	})
	contentPathEndpoint := app.String(cli.StringOpt{
		Name:   "contentPathEndpoint",
		Value:  "/content",
		Desc:   "/content path",
		EnvVar: "CONTENT_PATH",
	})
	internalContentPathEndpoint := app.String(cli.StringOpt{
		Name:   "internalContentPathEndpoint",
		Value:  "/internalcontent",
		Desc:   "/internalcontent path",
		EnvVar: "INTERNAL_CONTENT_PATH",
	})
	apiHost := app.String(cli.StringOpt{
		Name:   "apiHost",
		Value:  "test.api.ft.com",
		Desc:   "API host to use for URLs in responses",
		EnvVar: "API_HOST",
	})
	logLevel := app.String(cli.StringOpt{
		Name:   "logLevel",
		Value:  "INFO",
		Desc:   "Log level",
		EnvVar: "LOG_LEVEL",
	})

	log := logger.NewUPPLogger(AppName, *logLevel)

	app.Action = func() {
		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 100,
				DialContext: (&net.Dialer{
					KeepAlive: 30 * time.Second,
				}).DialContext,
			},
		}

		sc := content.ServiceConfig{
			ContentStoreAppName:      *contentStoreApplicationName,
			ContentStoreAppHealthURI: getServiceHealthURI(*contentStoreHost),
			HTTPClient:               httpClient,
		}

		readerConfig := content.ReaderConfig{
			ContentStoreAppName:         *contentStoreApplicationName,
			ContentStoreHost:            *contentStoreHost,
			ContentPathEndpoint:         *contentPathEndpoint,
			InternalContentPathEndpoint: *internalContentPathEndpoint,
		}

		reader := content.NewContentReader(readerConfig, httpClient)
		unroller := content.NewUniversalUnroller(reader, log, *apiHost)
		handler := content.NewHandler(unroller, log)

		h := setupServiceHandler(sc, *handler)
		err := http.ListenAndServe(":"+*port, h)
		if err != nil {
			log.Fatalf("Unable to start server: %v", err)
		}
	}

	log.Infof("Application started with args %s", os.Args)
	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Unable to start application: %v", err)
		return
	}
}

func setupServiceHandler(sc content.ServiceConfig, handler content.Handler) *mux.Router {
	r := mux.NewRouter()

	var checks []fthealth.Check
	var gtgHandler func(http.ResponseWriter, *http.Request)

	r.HandleFunc("/content", handler.GetContent).Methods("POST")
	r.HandleFunc("/internalcontent", handler.GetInternalContent).Methods("POST")
	checks = []fthealth.Check{sc.ContentStoreCheck()}
	gtgHandler = httphandlers.NewGoodToGoHandler(sc.GtgCheck)

	r.Path(httphandlers.BuildInfoPath).HandlerFunc(httphandlers.BuildInfoHandler)
	r.Path(httphandlers.PingPath).HandlerFunc(httphandlers.PingHandler)

	hc := fthealth.TimedHealthCheck{
		HealthCheck: fthealth.HealthCheck{SystemCode: AppCode, Name: AppName, Description: AppDesc, Checks: checks},
		Timeout:     10 * time.Second,
	}

	r.Path("/__health").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(fthealth.Handler(&hc))})
	r.Path("/__gtg").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(gtgHandler)})
	return r
}

func getServiceHealthURI(hostname string) string {
	return fmt.Sprintf("%s%s", hostname, "/__health")
}

//TODO: Check clipset arrays
