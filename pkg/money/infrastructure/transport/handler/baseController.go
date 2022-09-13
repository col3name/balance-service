package handler

import (
	"encoding/json"
	"github.com/col3name/balance-transfer/pkg/common/infrastructure/http/types"
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
	w.Header().Set(types.ContentType, types.ApplicationJson)
	w.WriteHeader(responseError.Status)
	_ = json.NewEncoder(w).Encode(responseError.Response)
}

func (c *BaseController) SetupCors(w *http.ResponseWriter, _ *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func (c *BaseController) WriteJsonResponse(writer http.ResponseWriter, data interface{}) {
	writer.Header().Set(types.ContentType, types.ApplicationJson)
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
