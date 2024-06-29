package handlers

import (
	"net/http"

	"github.com/jwtly10/jambda/pkg/logging"
)

type FunctionHandler struct {
	log logging.Logger
}

func NewFunctionHandler(l logging.Logger) *FunctionHandler {
	return &FunctionHandler{
		log: l,
	}
}

func (nfh *FunctionHandler) CreateNewFunction(w http.ResponseWriter, r *http.Request) {
	nfh.log.Info("Made a new request")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello world"))
}
