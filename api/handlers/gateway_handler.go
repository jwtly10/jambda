package handlers

import (
	"github.com/jwtly10/jambda/internal/logging"
	"github.com/jwtly10/jambda/internal/service"
)

type GatewayHandler struct {
	log     logging.Logger
	service service.GatewayService
}

func NewGatewayHandler(l logging.Logger, gs service.GatewayService) *GatewayHandler {
	return &GatewayHandler{
		log:     l,
		service: gs,
	}
}
