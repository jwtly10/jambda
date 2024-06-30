package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/jwtly10/jambda/api/data"
	"github.com/jwtly10/jambda/internal/logging"
	"github.com/jwtly10/jambda/internal/repository"
)

type DockerService struct {
	log logging.Logger
	fr  repository.FunctionRepository
	cli *client.Client
}

func NewDockerService(log logging.Logger, fr repository.FunctionRepository) *DockerService {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("failed to create docker client", err)
	}

	return &DockerService{
		log: log,
		cli: cli,
		fr:  fr,
	}
}

func (ds *DockerService) GetFunctionConfiguration(externalId string) (*data.FunctionConfig, error) {
	return ds.fr.GetConfigurationFromExternalId(externalId)
}

func (ds *DockerService) StartContainer(ctx context.Context, r *http.Request, functionId string, config data.FunctionConfig) (string, error) {
	// get list of all containers
	containers, err := ds.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		ds.log.Error("Failed to list Docker containers: ", err)
		return "", err
	}

	// get the containerId for either the running container, or the created container
	var containerId string
	containerFound := false
	for _, inContainer := range containers {
		if inContainer.Labels["function_id"] == functionId {
			containerId = inContainer.ID
			containerFound = true
			if inContainer.State == "running" {
				ds.log.Infof("Container for function %s is already running", functionId)
			} else {
				ds.log.Infof("Container for function %s exists but is not running. Starting it now.", functionId)
				// Start the container
				if err := ds.cli.ContainerStart(ctx, containerId, container.StartOptions{}); err != nil {
					ds.log.Error("Failed to start container: ", err)
					return "", err

				}
				// TODO. Implement health check here
				ds.log.Infof("TODO: Simulating health check")
				time.Sleep(2 * time.Second)
			}
			break
		}
	}

	if !containerFound {
		ds.log.Infof("No container found for id %s. Creating one now.", functionId)
		// TODO dont hard code path
		binaryPath := fmt.Sprintf("%s/binaries/%s/bootstrap", "/Users/personal/Projects/jambda", functionId)

		// Create and start the container
		cInstance, err := ds.cli.ContainerCreate(ctx, &container.Config{
			Image: config.Image,
			// TODO: Allow custom cmd params?
			Cmd: []string{"/bootstrap"},
			Labels: map[string]string{
				"function_id": functionId,
			},
			ExposedPorts: nat.PortSet{
				nat.Port(fmt.Sprintf("%d/tcp", *config.Port)): {},
			},
		}, &container.HostConfig{
			Binds: []string{
				binaryPath + ":/bootstrap:ro",
			},
			PortBindings: nat.PortMap{
				nat.Port(fmt.Sprintf("%d/tcp", *config.Port)): []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: "",
					},
				},
			},
		}, nil, nil, "")
		if err != nil {
			ds.log.Error("Failed to create container: ", err)
			return "", err
		}

		// Container was not found, so we created one...
		containerId = cInstance.ID

		// Start the container
		if err := ds.cli.ContainerStart(ctx, cInstance.ID, container.StartOptions{}); err != nil {
			ds.log.Error("Failed to start container: ", err)
			return "", err
		}

		// TODO. Implement health check here
		ds.log.Infof("TODO: Simulating health check")
		time.Sleep(2 * time.Second)
	}

	return containerId, nil
}

func (ds *DockerService) HealthCheckContainer(ctx context.Context, containerId string, config data.FunctionConfig) error {
	// Define a timeout for how long to wait for the container to become ready
	timeout := time.Now().Add(30 * time.Second) // Wait up to 30 seconds

	url, err := ds.GetContainerUrl(ctx, containerId, config)
	if err != nil {
		return err
	}

	for time.Now().Before(timeout) {
		resp, err := http.Get(fmt.Sprintf("%s/health", url))
		if err == nil && resp.StatusCode == http.StatusOK {
			return nil // Container is ready
		}
		// Wait a bit before checking again
		ds.log.Errorf("Error during health check: %v", err)
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("container did not become ready within the expected time")
}

func (ds *DockerService) GetContainerUrl(ctx context.Context, containerId string, config data.FunctionConfig) (string, error) {
	inspectData, err := ds.cli.ContainerInspect(ctx, containerId)
	if err != nil {
		ds.log.Error("Error inspecting container: ", err)
		return "", err
	}

	assignedPort := inspectData.NetworkSettings.Ports[nat.Port(fmt.Sprintf("%d/tcp", *config.Port))][0].HostPort
	return fmt.Sprintf("http://%s:%s", "localhost", assignedPort), nil
}
