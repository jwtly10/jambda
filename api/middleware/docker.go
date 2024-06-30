package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/jwtly10/jambda/internal/logging"
	"github.com/jwtly10/jambda/internal/service"
)

type DockerMiddleware struct {
	Log logging.Logger
	Ds  service.DockerService
}

func (dmw *DockerMiddleware) BeforeNext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the functionId from the request
		functionId := strings.TrimPrefix(r.URL.Path, "/v1/api/function/")
		functionId = strings.SplitN(functionId, "/", 2)[0]

		// 1. Validate the function id by getting the config for the function
		config, err := dmw.Ds.GetFunctionConfiguration(functionId)
		if err != nil {
			dmw.Log.Errorf("failed to get function config for id: %s %v", functionId, err)
			// TODO FIX ERROR MESSAGING
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return

		}

		ctx := context.Background()

		// 2. Run functions based on function type
		switch funcType := config.Type; funcType {
		case "REST":
			// This is a rest server function. So we should spin up the server in a container
			containerId, err := dmw.Ds.StartContainer(ctx, r, functionId, *config)
			if err != nil {
				dmw.Log.Errorf("Error starting container: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Inject the ip and port so we can check health
			// Pass the right ip and port of the container to the handler to route request properly
			containerUrl, err := dmw.Ds.GetContainerUrl(ctx, containerId, *config)
			if err != nil {
				return
			}

			// TODO we should run health check, for now just wait a few seconds
			// The current issue is we cannot get any ports unit service is running. So we need to
			// Spend a few tries to get the port/url
			// dmw.Log.Infof("Running health check!!")
			// err = dmw.Ds.HealthCheckContainer(ctx, containerId, *config)
			// if err != nil {
			// 	dmw.Log.Errorf("%v", err)
			// 	http.Error(w, "Internal server error", http.StatusInternalServerError)
			// 	return
			// }

			// Pass everything to the handler, including the url of the running container
			dmw.Log.Infof("Determined container '%s' url : '%s", containerId, containerUrl)
			r = r.WithContext(context.WithValue(r.Context(), "containerUrl", containerUrl))
			next.ServeHTTP(w, r)
		case "SINGLE":
			http.Error(w, "Not implemented yet", http.StatusNotImplemented)
			return
		default:
			http.Error(w, "Unsupported function type", http.StatusBadRequest)
			return
		}

	})
}
