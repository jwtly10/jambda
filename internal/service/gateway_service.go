package service

import "github.com/jwtly10/jambda/internal/logging"

type GatewayService struct {
	log logging.Logger
}

func NewGatewayService(log logging.Logger) *GatewayService {
	return &GatewayService{
		log: log,
	}
}

func (gs *GatewayService) DoSomething() {
}
