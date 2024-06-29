package api

import (
	"fmt"
	"net/http"

	"github.com/jwtly10/jambda/api/middleware"
	"github.com/jwtly10/jambda/pkg/logging"
)

type AppRouter struct {
	*http.ServeMux
	middleware []middleware.Middleware
	logger     logging.Logger
}

func NewAppRouter(l logging.Logger) AppRouter {
	return AppRouter{
		logger:   l,
		ServeMux: http.NewServeMux(),
	}
}

func (r *AppRouter) handle(pattern string, handler http.Handler) {
	for _, middleware := range r.middleware {
		handler = middleware.BeforeNext(handler)
	}
	r.ServeMux.Handle(pattern, handler)
}

func (r *AppRouter) Use(middlewares ...middleware.Middleware) {
	r.middleware = append(r.middleware, middlewares...)
}

func (r *AppRouter) Get(pattern string, handler http.Handler) {
	r.logger.Infof("Mapped [GET] %s", pattern)
	r.handle(fmt.Sprintf("GET %s", pattern), handler)
}

func (r *AppRouter) Post(pattern string, handler http.Handler) {
	r.logger.Infof("Mapped [POST] %s", pattern)
	r.handle(fmt.Sprintf("POST %s", pattern), handler)
}

func (r *AppRouter) Update(pattern string, handler http.Handler) {
	r.logger.Infof("Mapped [UPDATE] %s", pattern)
	r.handle(fmt.Sprintf("UPDATE %s", pattern), handler)
}

func (r *AppRouter) Put(pattern string, handler http.Handler) {
	r.logger.Infof("Mapped [PUT] %s", pattern)
	r.handle(fmt.Sprintf("PUT %s", pattern), handler)
}

func (r *AppRouter) Delete(pattern string, handler http.Handler) {
	r.logger.Infof("Mapped [DELETE] %s", pattern)
	r.handle(fmt.Sprintf("DELETE %s", pattern), handler)
}

func (r *AppRouter) Options(pattern string, handler http.Handler) {
	r.logger.Infof("Mapped [OPTIONS] %s", pattern)
	r.handle(fmt.Sprintf("OPTIONS %s", pattern), handler)
}
