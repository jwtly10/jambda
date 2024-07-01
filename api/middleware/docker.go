package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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

			containerUrl, err := dmw.Ds.GetContainerUrl(ctx, containerId, *config)
			if err != nil {
				utils.HandleCustomErrors(w, err)
				return
			}

			// TODO we should run health check on this container, for now just wait a few seconds
			// The current issue is we cannot get any ports until service is running. So we need to
			// Spend a few tries to get the port/url
			// dmw.Log.Infof("Running health check!!")
			// err = dmw.Ds.HealthCheckContainer(ctx, containerId, *config)
			// if err != nil {
			// 	dmw.Log.Errorf("%v", err)
			// 	http.Error(w, "Internal server error", http.StatusInternalServerError)
			// 	return
			// }

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
