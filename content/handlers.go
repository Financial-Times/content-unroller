package content

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Financial-Times/go-logger/v2"
	transactionidutils "github.com/Financial-Times/transactionid-utils-go"
)

type Unroller interface {
	UnrollContent(event UnrollEvent) (Content, error)
	UnrollInternalContent(event UnrollEvent) (Content, error)
}

type Handler struct {
	Unroller Unroller
	log      *logger.UPPLogger
}

func NewHandler(u Unroller, l *logger.UPPLogger) *Handler {
	return &Handler{Unroller: u, log: l}
}

type UnrollEvent struct {
	c    Content
	tid  string
	uuid string
}

func (hh *Handler) GetContent(w http.ResponseWriter, r *http.Request) {
	tid := transactionidutils.GetTransactionIDFromRequest(r)
	event, err := createUnrollEvent(r, tid)
	if err != nil {
		handleError(r, hh.log, tid, "", w, err, http.StatusBadRequest)
		return
	}

	transactionStartedEvent(hh.log, r.RequestURI, tid, event.uuid)

	res, err := hh.Unroller.UnrollContent(event)
	if errors.Is(err, ErrConnectingToAPI) {
		handleError(r, hh.log, tid, event.uuid, w, err, http.StatusInternalServerError)
		return
	} else if errors.Is(err, ErrValidating) {
		handleError(r, hh.log, tid, event.uuid, w, err, http.StatusBadRequest)
		return
	} else if err != nil {
		handleError(r, hh.log, tid, event.uuid, w, err, http.StatusInternalServerError)
		return
	}

	jsonRes, err := json.Marshal(res)
	if err != nil {
		handleError(r, hh.log, tid, event.uuid, w, err, http.StatusInternalServerError)
		return
	}

	transactionFinishedEvent(hh.log, r.RequestURI, tid, http.StatusOK, event.uuid, "success")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(jsonRes)
}

func (hh *Handler) GetInternalContent(w http.ResponseWriter, r *http.Request) {
	tid := transactionidutils.GetTransactionIDFromRequest(r)
	event, err := createUnrollEvent(r, tid)
	if err != nil {
		handleError(r, hh.log, tid, "", w, err, http.StatusBadRequest)
	}

	transactionStartedEvent(hh.log, r.RequestURI, tid, event.uuid)

	res, err := hh.Unroller.UnrollInternalContent(event)
	if errors.Is(err, ErrConnectingToAPI) {
		handleError(r, hh.log, tid, event.uuid, w, err, http.StatusInternalServerError)
		return
	} else if errors.Is(err, ErrValidating) {
		handleError(r, hh.log, tid, event.uuid, w, err, http.StatusBadRequest)
		return
	} else if err != nil {
		handleError(r, hh.log, tid, event.uuid, w, err, http.StatusInternalServerError)
		return
	}

	jsonRes, err := json.Marshal(res)
	if err != nil {
		handleError(r, hh.log, tid, event.uuid, w, err, http.StatusInternalServerError)
		return
	}

	transactionFinishedEvent(hh.log, r.RequestURI, tid, http.StatusOK, event.uuid, "success")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(jsonRes)
}

func createUnrollEvent(r *http.Request, tid string) (UnrollEvent, error) {
	var unrollEvent UnrollEvent
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return unrollEvent, err
	}

	var content Content
	err = json.Unmarshal(b, &content)
	if err != nil {
		return unrollEvent, err
	}

	//TODO: This may need to be moved to a validation function in the unroller in case `id` is not present in any of the unrollable content
	id, ok := content[id].(string)
	if !ok {
		return unrollEvent, errors.New("Missing or invalid id field")
	}
	uuid, err := extractUUIDFromString(id)
	if err != nil {
		return unrollEvent, err
	}
	unrollEvent = UnrollEvent{content, tid, uuid}

	return unrollEvent, nil
}

func handleError(r *http.Request, log *logger.UPPLogger, tid string, uuid string, w http.ResponseWriter, err error, statusCode int) {
	var errMsg string
	if statusCode >= 400 && statusCode < 500 {
		errMsg = fmt.Sprintf("Error expanding content, supplied UUID is invalid: %s", err.Error())
		transactionFinishedEvent(log, r.RequestURI, tid, statusCode, uuid, err.Error())
	} else if statusCode >= 500 {
		errMsg = fmt.Sprintf("Error expanding content for: %v: %v", uuid, err.Error())
		transactionFinishedEvent(log, r.RequestURI, tid, statusCode, uuid, err.Error())
	}
	w.WriteHeader(statusCode)
	w.Write([]byte(errMsg))
}
