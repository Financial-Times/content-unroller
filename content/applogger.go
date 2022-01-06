package content

import (
	"github.com/sirupsen/logrus"
)

type appLogger struct {
	log *logrus.Logger
}

func NewAppLogger() *appLogger {
	logrus.SetLevel(logrus.InfoLevel)
	log := logrus.New()
	log.Formatter = new(logrus.JSONFormatter)
	return &appLogger{log: log}
}

func (appLogger *appLogger) TransactionStartedEvent(requestURL string, transactionID string, uuid string) {
	appLogger.log.WithFields(logrus.Fields{
		"request_url":    requestURL,
		"transaction_id": transactionID,
		"uuid":           uuid,
	}).Infof("Transaction started %s", transactionID)
}

func (appLogger *appLogger) TransactionFinishedEvent(requestURL string, transactionID string, statusCode int, uuid string, message string) {
	e := appLogger.log.WithFields(logrus.Fields{
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

func (appLogger *appLogger) Debugf(tid string, uuid string, format string, args ...interface{}) {
	appLogger.log.WithFields(logrus.Fields{"tid": tid, "uuid": uuid}).Debugf(format, args...)
}

func (appLogger *appLogger) Debug(tid string, uuid string, message string) {
	appLogger.log.WithFields(logrus.Fields{"tid": tid, "uuid": uuid}).Debug(message)
}

func (appLogger *appLogger) Infof(tid string, uuid string, format string, args ...interface{}) {
	appLogger.log.WithFields(logrus.Fields{"tid": tid, "uuid": uuid}).Infof(format, args...)
}

func (appLogger *appLogger) Info(tid string, uuid string, message string) {
	appLogger.log.WithFields(logrus.Fields{"tid": tid, "uuid": uuid}).Info(message)
}

func (appLogger *appLogger) Errorf(tid string, format string, args ...interface{}) {
	appLogger.log.WithFields(logrus.Fields{"tid": tid}).Errorf(format, args...)
}

func (appLogger *appLogger) Warnf(tid string, uuid string, format string, args ...interface{}) {
	appLogger.log.WithFields(logrus.Fields{"tid": tid, "uuid": uuid}).Warnf(format, args...)
}
