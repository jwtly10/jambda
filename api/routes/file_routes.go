package routes

import (
	"net/http"

	"github.com/jwtly10/jambda/api"
	"github.com/jwtly10/jambda/api/handlers"
	"github.com/jwtly10/jambda/api/middleware"
	"github.com/jwtly10/jambda/internal/logging"
)

type FileRoutes struct {
	log      logging.Logger
	handlers handlers.FileHandler
}

func NewFileRoutes(router api.AppRouter, l logging.Logger, h handlers.FileHandler, mws ...middleware.Middleware) FileRoutes {
	routes := FileRoutes{
		log:      l,
		handlers: h,
	}

	BASE_PATH := "/v1/api"

	uploadHandler := http.HandlerFunc(routes.handlers.UploadFile)
	router.Post(
		BASE_PATH+"/file/upload",
		middleware.Chain(uploadHandler, mws...),
	)

	return routes
}
