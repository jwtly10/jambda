package middleware

import (
	"fmt"
	"net/http"

	"github.com/jwtly10/jambda/internal/logging"
	"github.com/jwtly10/jambda/internal/service"
	"github.com/jwtly10/jambda/internal/utils"
)

type UsageMiddleware struct {
	log logging.Logger
	rs  *service.RequestStatsService
}

func NewUsageMiddleware(log logging.Logger, rs *service.RequestStatsService) *UsageMiddleware {
	return &UsageMiddleware{
		log: log,
		rs:  rs,
	}
}

func (umw *UsageMiddleware) BeforeNext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the function ID from the URL
		functionId := utils.GetFunctionIdFromExecutePath(r)
		if functionId == "" {
			utils.HandleValidationError(w, fmt.Errorf("function ID is required"))
		}

		umw.log.Infof("Incrementing request count for %s", functionId)
		umw.rs.IncrementRequestCount(functionId)

		next.ServeHTTP(w, r)
	})
}
