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

// CreateNewFunction serves a basic hello world message
// @Summary Serve hello world
// @Description Responds with a plain text "Hello world" to any request.
// @Tags functions
// @Accept  */*
// @Produce plain
// @Success 200 {string} string "Hello world"
// @Router /api/upload-binary [get]
func (nfh *FunctionHandler) CreateNewFunction(w http.ResponseWriter, r *http.Request) {
	nfh.log.Info("Made a new request")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello world"))
}
