package routes

import (
	"net/http"

	"github.com/jwtly10/jambda/api"
	"github.com/jwtly10/jambda/api/handlers"
	"github.com/jwtly10/jambda/api/middleware"
	"github.com/jwtly10/jambda/internal/logging"
)

type FunctionRoutes struct {
	log      logging.Logger
	handlers handlers.FunctionHandler
}

func NewFunctionRoutes(router api.AppRouter, l logging.Logger, h handlers.FunctionHandler, mws ...middleware.Middleware) FunctionRoutes {
	routes := FunctionRoutes{
		log:      l,
		handlers: h,
	}

	BASE_PATH := "/v1/api"

	uploadHandler := http.HandlerFunc(routes.handlers.UploadFunction)
	router.Post(
		BASE_PATH+"/function",
		middleware.Chain(uploadHandler, mws...),
	)

	updateHandler := http.HandlerFunc(routes.handlers.UpdateFunction)
	router.Put(
		BASE_PATH+"/function/{id}",
		middleware.Chain(updateHandler, mws...),
	)

	listHandler := http.HandlerFunc(routes.handlers.ListFunctions)
	router.Get(
		BASE_PATH+"/function",
		middleware.Chain(listHandler, mws...),
	)

	deleteHandler := http.HandlerFunc(routes.handlers.DeleteFunction)
	router.Delete(
		BASE_PATH+"/function/{id}",
		middleware.Chain(deleteHandler, mws...),
	)
	return routes
}
