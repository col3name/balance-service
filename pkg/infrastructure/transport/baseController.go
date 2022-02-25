package transport

import (
	"encoding/json"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var (
	ErrBadRouting = errors.New("bad routing")
	ErrBadRequest = errors.New("bad request")
)

type BaseController struct {
}

func (c *BaseController) WriteError(w http.ResponseWriter, err error, responseError *Error) {
	log.Error(err.Error())
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(responseError.Status)
	_ = json.NewEncoder(w).Encode(responseError.Response)
}

func (c *BaseController) SetupCors(w *http.ResponseWriter, _ *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func (c *BaseController) WriteJsonResponse(writer http.ResponseWriter, data interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(jsonData)
	if err != nil {
		return
	}
}
