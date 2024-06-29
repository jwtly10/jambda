package routes

import (
	"net/http"

	"github.com/jwtly10/jambda/api"
	"github.com/jwtly10/jambda/api/handlers"
	"github.com/jwtly10/jambda/api/middleware"
	"github.com/jwtly10/jambda/pkg/logging"
)

type UploadRoutes struct {
	log      logging.Logger
	handlers handlers.FunctionHandler
}

func NewUploadRoutes(router api.AppRouter, l logging.Logger, h handlers.FunctionHandler, mws ...middleware.Middleware) UploadRoutes {
	routes := UploadRoutes{
		log:      l,
		handlers: h,
	}

	uploadHandler := http.HandlerFunc(routes.handlers.CreateNewFunction)
	router.Get(
		"/api/upload",
		middleware.Chain(uploadHandler, mws...),
	)

	return routes
}
