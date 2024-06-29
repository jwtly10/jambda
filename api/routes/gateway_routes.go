package routes

import (
	"github.com/jwtly10/jambda/api"
	"github.com/jwtly10/jambda/api/handlers"
	"github.com/jwtly10/jambda/api/middleware"
	"github.com/jwtly10/jambda/internal/logging"
)

type GatewayRoutes struct {
	log      logging.Logger
	handlers handlers.GatewayHandler
}

func NewGatewayRoutes(router api.AppRouter, l logging.Logger, h handlers.GatewayHandler, mws ...middleware.Middleware) GatewayRoutes {
	routes := GatewayRoutes{
		log:      l,
		handlers: h,
	}

	// BASE_PATH := "/v1/api"

	// router.Post(
	// 	BASE_PATH+"/file/upload",
	// 	middleware.Chain(uploadHandler, mws...),
	// )

	return routes
}
