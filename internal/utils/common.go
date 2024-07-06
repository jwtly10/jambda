package utils

import (
	"net/http"
	"strings"
)

func GetFunctionIdFromExecutePath(r *http.Request) string {
	// Get the functionId from the request
	functionId := strings.TrimPrefix(r.URL.Path, "/v1/api/execute/")
	functionId = strings.SplitN(functionId, "/", 2)[0]
	return functionId
}
