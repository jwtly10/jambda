package handlers

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/jwtly10/jambda/internal/logging"
	"github.com/jwtly10/jambda/internal/service"
	"github.com/jwtly10/jambda/internal/utils"
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

// @Summary Make request to a REST function
// @Description Proxies requests to docker instance running executable. Method passed to instance forwarded from req. Middleware figures out the instance URL to proxy the request to, based on ExternalId. Returns proxied response.
// @Tags Executions
// @Accept plain
// @Produce */*
// @Param id path string true "External ID"
// @Success 200 {string} string "Request successfully proxied and processed"
// @Failure 400 {object} utils.ErrorResponse "Bad Request"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /execute/{id}/ [post]
// @Router /execute/{id}/ [get]
// @Router /execute/{id}/ [put]
// @Router /execute/{id}/ [delete]
func (gwh *GatewayHandler) ProxyToInstance(w http.ResponseWriter, r *http.Request) {
	// Retrieve the base URL from the context, set by docker middleware
	baseContainerUrl, ok := r.Context().Value("containerUrl").(string)
	if !ok || baseContainerUrl == "" {
		utils.HandleInternalError(w, fmt.Errorf("Container URL not passed through context"))
		return
	}

	url, err := parseProxiedUrlGivenBaseUrl(baseContainerUrl, r.URL)
	if err != nil {
		gwh.log.Errorf("Unable to route request to container: %v", err)
		utils.HandleInternalError(w, fmt.Errorf("Failed to parse url and proxy request - %v", err))
		return
	}

	gwh.log.Infof("Proxing request to instance url: '%s'", url)

	// Configure reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.Director = func(req *http.Request) {
		req.URL = url
		req.Host = url.Host
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Header.Set("X-Real-IP", req.RemoteAddr)
	}

	// Serve HTTP through the proxy
	proxy.ServeHTTP(w, r)

}

func parseProxiedUrlGivenBaseUrl(baseUrl string, proxiedUrl *url.URL) (*url.URL, error) {
	// NOTE!!! This assumes the pattern is /v1/api/execute/{id}/extra?param=params
	pathParts := strings.SplitN(proxiedUrl.Path, "/", 6)
	if len(pathParts) < 5 {
		return nil, fmt.Errorf("invalid request path")
	}

	// Rebuild the path that needs to be forwarded to
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
