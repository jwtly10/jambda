package routes

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

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
		// Retrieve the base URL from the context, set by middleware
		baseContainerUrl, ok := r.Context().Value("containerUrl").(string)
		if !ok || baseContainerUrl == "" {
			http.Error(w, "Container URL not found", http.StatusInternalServerError)
			return
		}

		url, err := parseProxiedUrlGivenBaseUrl(baseContainerUrl, r.URL)
		if err != nil {
			routes.log.Errorf("Unable to route request to container: %v", err)
			http.Error(w, "Unable to route request to container", http.StatusBadRequest)
			return
		}

		// Create and configure the reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.Director = func(req *http.Request) {
			req.URL = url
			req.Host = url.Host
			req.Header.Set("X-Forwarded-Host", req.Host)
			req.Header.Set("X-Real-IP", req.RemoteAddr)
		}

		// Serve HTTP through the proxy
		proxy.ServeHTTP(w, r)
	})

	router.Post(
		BASE_PATH+"/function/{id}/*",
		middleware.Chain(gatewayHandler, mws...),
	)

	return routes
}

func parseProxiedUrlGivenBaseUrl(baseUrl string, proxiedUrl *url.URL) (*url.URL, error) {
	pathParts := strings.SplitN(proxiedUrl.Path, "/", 6) // Assuming the pattern is /v1/api/function/{id}/extra
	if len(pathParts) < 5 {
		return nil, fmt.Errorf("invalid request path")
	}
	// Rebuild the path that needs to be forwarded
	forwardPath := "/" + strings.Join(pathParts[5:], "/") + "?" + proxiedUrl.RawQuery

	// Remove trailing ?
	if forwardPath[len(forwardPath)-1:] == "?" {
		forwardPath = forwardPath[0 : len(forwardPath)-1]
	}

	// Construct the full URL to forward the request to
	destinationUrl := baseUrl + forwardPath

	// Parse the destination URL
	url, err := url.Parse(destinationUrl)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse destination Url")
	}

	return url, nil
}
