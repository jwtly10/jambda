package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jwtly10/jambda/internal/logging"
	"github.com/jwtly10/jambda/internal/service"
	"github.com/jwtly10/jambda/internal/utils"
)

type DockerMiddleware struct {
	Log logging.Logger
	Ds  service.DockerService
}

func (dmw *DockerMiddleware) BeforeNext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the functionId from the request
		functionId := strings.TrimPrefix(r.URL.Path, "/v1/api/execute/")
		functionId = strings.SplitN(functionId, "/", 2)[0]

		// 1. Validate the function id by getting the config for the function
		config, err := dmw.Ds.GetFunctionConfiguration(functionId)
		if err != nil {
			dmw.Log.Errorf("Failed to get function config for id: %s %v", functionId, err)
			utils.HandleCustomErrors(w, err)
			return
		}

		if config.Trigger != "http" {
			dmw.Log.Errorf("Unsupported http trigger: '%s'", config.Trigger)
			utils.HandleValidationError(w, fmt.Errorf("Function config trigger '%s' is not supported", config.Trigger))
			return
		}

		ctx := context.Background()

		// 2. Run functions based on function type
		switch funcType := config.Type; funcType {
		case "REST":
			containerId, err := dmw.Ds.StartContainer(ctx, r, functionId, *config)
			if err != nil {
				dmw.Log.Errorf("Error starting container: %v", err)
				utils.HandleCustomErrors(w, err)
				return
			}

			// We need to get the url from the container that has been started (or is already running)
			// However, it can take some time for docker to spin up the process, so we don't want to proxy the request
			// If the container is not running
			// This logic tries for max 5 seconds to check the container is running.
			var containerUrl string
			timeout := time.Now().Add(5 * time.Second)
			for time.Now().Before(timeout) {
				containerUrl, err = dmw.Ds.GetContainerUrl(ctx, containerId, *config)
				if err == nil {
					// If there is no error, that means we got the container url
					// So we can continue with the health check
					break
				}
				dmw.Log.Warn("Url check failed. Retrying.")
				time.Sleep(2 * time.Second)
			}
			if err != nil {
				// We didn't get the container URL!
				dmw.Log.Error("Error getting container URL : %v", err)
				utils.HandleInternalError(w, fmt.Errorf("unable to get container url! The container must not have been ready in time %v", err))
				return
			}

			// Otherwise we can continue with the health check
			// This also attempts to wait, and retry the health check.
			// As this is not a problem within docker, but the underlying function.
			// We will wait up to 30 seconds
			// Rest platforms like springboot can take a few seconds to initialise.
			dmw.Log.Infof("Running health check!!")
			err = dmw.Ds.HealthCheckContainer(ctx, containerId, *config)
			if err != nil {
				dmw.Log.Errorf("Health check failed %v", err)
				utils.HandleCustomErrors(w, err)
				return
			}

			// Pass everything to the handler, including the url of the running container
			dmw.Log.Infof("Determined container '%s' url : '%s'", containerId, containerUrl)
			r = r.WithContext(context.WithValue(r.Context(), "containerUrl", containerUrl))
			next.ServeHTTP(w, r)
		case "SINGLE":
			// Here we will instead execute the binary, and if it can run within a few seconds
			// Return the sdout
			// Else just respond with running, maybe in future provide some option to notify when complete
			http.Error(w, "Not implemented yet", http.StatusNotImplemented)
			return
		default:
			utils.HandleBadRequest(w, fmt.Errorf("Unsupported function type '%s'", config.Type))
			http.Error(w, "Unsupported function type", http.StatusBadRequest)
			return
		}
	})
}
