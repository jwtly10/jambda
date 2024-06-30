package routes

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

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

	BASE_PATH := "/v1/api"

	gatewayHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		containerIP, ok := r.Context().Value("containerIP").(string)
		if !ok || containerIP == "" {
			http.Error(w, "Container IP not found", http.StatusInternalServerError)
			return
		}
		// containerPort, ok := r.Context().Value("containerPort").(int)
		// if !ok || containerIP == "" {
		// 	http.Error(w, "Container port not found", http.StatusInternalServerError)
		// 	return
		// }

		// Construct the destination URL with the retrieved IP address
		destURL := fmt.Sprintf("http://%s:%d", "localhost", 8001) // Port needs to be known
		routes.log.Infof("Container destination url: %s", destURL)

		// Parse the destination URL
		url, err := url.Parse(destURL)
		if err != nil {
			http.Error(w, "Failed to parse destination URL", http.StatusInternalServerError)
			return
		}

		// Create the reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.ServeHTTP(w, r)
	})

	router.Post(
		BASE_PATH+"/function/{id}/*",
		middleware.Chain(gatewayHandler, mws...),
	)

	return routes
}
