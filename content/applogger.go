package content

import "github.com/Financial-Times/go-logger/v2"

func transactionStartedEvent(log *logger.UPPLogger, requestURL string, transactionID string, uuid string) {
	log.WithFields(map[string]interface{}{
		"request_url":    requestURL,
		"transaction_id": transactionID,
		"uuid":           uuid,
	}).Infof("Transaction started %s", transactionID)
}

func transactionFinishedEvent(log *logger.UPPLogger, requestURL string, transactionID string, statusCode int, uuid string, message string) {
	e := log.WithFields(map[string]interface{}{
		"request_url":    requestURL,
		"transaction_id": transactionID,
		"uuid":           uuid,
	})

	if statusCode < 300 {
		e.Infof("Transaction %s finished with status %d: %s", transactionID, statusCode, message)
	} else {
		e.Errorf("Transaction %s finished with status %d: %s", transactionID, statusCode, message)
	}
}
