package middleware

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/jwtly10/jambda/internal/logging"
	"github.com/jwtly10/jambda/internal/utils"
)

// Struct to hold the request count
type FunctionStats struct {
	RequestCount int
}

// Global map and mutex to store function statistics
var (
	functionStats = make(map[string]*FunctionStats)
	mu            sync.Mutex
)

type UsageMiddleware struct {
	Log logging.Logger
}

func NewUsageMiddleware(log logging.Logger) *UsageMiddleware {
	return &UsageMiddleware{
		Log: log,
	}
}

func (umw *UsageMiddleware) BeforeNext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the function ID from the URL
		functionId := utils.GetFunctionIdFromExecutePath(r)
		if functionId == "" {
			utils.HandleValidationError(w, fmt.Errorf("function ID is required"))
		}

		umw.Log.Infof("Incrementing request count for %s", functionId)
		incrementRequestCount(functionId)

		next.ServeHTTP(w, r)
	})
}

func incrementRequestCount(functionID string) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := functionStats[functionID]; !exists {
		functionStats[functionID] = &FunctionStats{}
	}
	functionStats[functionID].RequestCount++
}
