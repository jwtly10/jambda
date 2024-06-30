package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/jwtly10/jambda/internal/logging"
	"github.com/jwtly10/jambda/internal/service"
)

type DockerMiddleware struct {
	Log logging.Logger
	Fs  service.FileService
}

func (dmw *DockerMiddleware) BeforeNext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the functionID from the request
		functionID := strings.TrimPrefix(r.URL.Path, "/v1/api/function/")
		functionID = strings.SplitN(functionID, "/", 2)[0]

		if !dmw.Fs.IsValidFunctionId(functionID) {
			dmw.Log.Errorf("failed to validate function id %s", functionID)
			http.Error(w, "Invalid function id", http.StatusNoContent)
			return
		}

		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			dmw.Log.Error("Failed to create Docker client: ", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Get all containers on system
		containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
		if err != nil {
			dmw.Log.Error("Failed to list Docker containers: ", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		var containerID string
		containerFound := false
		for _, inContainer := range containers {
			if inContainer.Labels["function_id"] == functionID {
				containerID = inContainer.ID
				containerFound = true
				if inContainer.State == "running" {
					dmw.Log.Infof("Container for function %s is already running", functionID)
				} else {
					dmw.Log.Infof("Container for function %s exists but is not running. Starting it now.", functionID)
					// Start the container
					if err := cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
						dmw.Log.Error("Failed to start container: ", err)
						http.Error(w, "Internal server error", http.StatusInternalServerError)
						return
					}
				}
				break
			}
		}

		if !containerFound {
			dmw.Log.Infof("No container found for id %s. Creating one now.", functionID)
			binaryPath := fmt.Sprintf("%s/binaries/%s/bootstrap", "/Users/personal/Projects/jambda", functionID) // Use an absolute path

			// Create and start the container
			cInstance, err := cli.ContainerCreate(ctx, &container.Config{
				Image: "golang:1.22",          // Ensure this image has the necessary runtime to execute your binary
				Cmd:   []string{"/bootstrap"}, // Command to run inside the container
				Labels: map[string]string{
					"function_id": functionID,
				},
				ExposedPorts: nat.PortSet{
					nat.Port("6969/tcp"): {},
				},
			}, &container.HostConfig{
				Binds: []string{
					binaryPath + ":/bootstrap:ro", // Mount the binary into the container
				},
				PortBindings: nat.PortMap{
					nat.Port("6969/tcp"): []nat.PortBinding{
						{
							HostIP:   "0.0.0.0",
							HostPort: "8001",
						},
					},
				},
			}, nil, nil, "")

			// update container id
			containerID = cInstance.ID

			if err != nil {
				dmw.Log.Error("Failed to create container: ", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Start the container
			if err := cli.ContainerStart(ctx, cInstance.ID, container.StartOptions{}); err != nil {
				dmw.Log.Error("Failed to start container: ", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}

		// At this point we can assume the container exists

		inspectData, err := cli.ContainerInspect(ctx, containerID)
		if err != nil {
			dmw.Log.Error("Failed to inspect container: ", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		dmw.Log.Infof("Port Bindings: %v", inspectData.HostConfig.PortBindings)

		for _, network := range inspectData.NetworkSettings.Networks {
			r = r.WithContext(context.WithValue(r.Context(), "containerIP", network.IPAddress))
			// Now, attempt to find the external port mapped to the internal port (if applicable)
			var externalPort string
			for port, bindings := range inspectData.NetworkSettings.Ports {
				dmw.Log.Infof("Port %v", port)
				dmw.Log.Infof("Bindings %v", bindings)
				if len(bindings) > 0 {
					externalPort = bindings[0].HostPort // Take the first binding (simplification)
					break
				}
			}
			r = r.WithContext(context.WithValue(r.Context(), "containerPort", externalPort))
			break
		}

		dmw.Log.Info("Waiting for health check")
		err = waitForContainerReady("http://localhost:8001")
		if err != nil {
			dmw.Log.Errorf("%v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return

		}

		next.ServeHTTP(w, r)
	})

}

func waitForContainerReady(url string) error {
	// Define a timeout for how long to wait for the container to become ready
	timeout := time.Now().Add(30 * time.Second) // Wait up to 30 seconds

	for time.Now().Before(timeout) {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			return nil // Container is ready
		}

		// Wait a bit before checking again
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("container did not become ready within the expected time")
}
