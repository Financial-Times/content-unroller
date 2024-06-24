package content

import (
	"fmt"
	"net/http"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/service-status-go/gtg"
)

type ServiceConfig struct {
	ContentStoreAppName      string
	ContentStoreAppHealthURI string
	HTTPClient               *http.Client
}

func (sc *ServiceConfig) GtgCheck() gtg.Status {
	contentStoreCheck := func() gtg.Status {
		msg, err := sc.checkServiceAvailability(sc.ContentStoreAppName, sc.ContentStoreAppHealthURI)
		if err != nil {
			return gtg.Status{GoodToGo: false, Message: msg}
		}
		return gtg.Status{GoodToGo: true}
	}
	return gtg.FailFastParallelCheck([]gtg.StatusChecker{
		contentStoreCheck,
	})()
}

func (sc *ServiceConfig) ContentStoreCheck() fthealth.Check {
	return fthealth.Check{
		ID:               fmt.Sprintf("check-connect-%s", sc.ContentStoreAppName),
		Name:             fmt.Sprintf("Check connectivity to %s", sc.ContentStoreAppName),
		Severity:         1,
		BusinessImpact:   "Unrolled images and dynamic content won't be available",
		TechnicalSummary: fmt.Sprintf(`Cannot connect to %v.`, sc.ContentStoreAppName),
		PanicGuide:       "https://dewey.in.ft.com/runbooks/contentreadapi",
		Checker: func() (string, error) {
			return sc.checkServiceAvailability(sc.ContentStoreAppName, sc.ContentStoreAppHealthURI)
		},
	}
}

func (sc *ServiceConfig) checkServiceAvailability(serviceName string, healthURI string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, healthURI, nil)
	if err != nil {
		return "Error", fmt.Errorf("%s URL cannot be used to create a request: %v", healthURI, err)
	}
	resp, err := sc.HTTPClient.Do(req)
	if err != nil {
		return "Error", fmt.Errorf("%s service is unreachable: %v", serviceName, err)
	}
	if resp.StatusCode != http.StatusOK {
		return "Error", fmt.Errorf("%s service is not responding with OK. Status=%d", serviceName, resp.StatusCode)
	}
	return "Ok", nil
}
