package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/jwtly10/jambda/api/data"
	"github.com/jwtly10/jambda/internal/errors"
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
	config, err := ds.fr.GetConfigurationFromExternalId(externalId)
	if err != nil {
		ds.log.Error("Failed to retrieve config from function '%s': ", externalId, err)
		return nil, errors.NewInternalError(fmt.Sprintf("error retrieving function config from db: %v", err))
	}

	return config, nil
}

func (ds *DockerService) StartContainer(ctx context.Context, r *http.Request, functionId string, config data.FunctionConfig) (string, error) {
	// get list of all containers
	containers, err := ds.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		ds.log.Error("Failed to list Docker containers: ", err)
		return "", errors.NewDockerError(fmt.Sprintf("error retrieving containers from docker: %v", err))
	}

	// get the containerId for either the running container, or the created container
	var containerId string
	containerFound := false
	for _, inContainer := range containers {
		if inContainer.Labels["function_id"] == functionId {
			containerId = inContainer.ID
			containerFound = true
			if inContainer.State == "running" {
				ds.log.Infof("Container for function '%s' is already running", functionId)
			} else {
				ds.log.Infof("Container for function '%s' exists but is not running. Starting it now.", functionId)
				// Start the container
				if err := ds.cli.ContainerStart(ctx, containerId, container.StartOptions{}); err != nil {
					ds.log.Error("Failed to start container '%s': ", containerId, err)
					return "", errors.NewDockerError(fmt.Sprintf("error starting docker container: %v", err))
				}
			}
			break
		}
	}

	if !containerFound {
		ds.log.Infof("No container found for id '%s'. Creating one now.", functionId)
		// TODO dont hard code path
		var runCmd []string
		var mountCmd string
		var binaryPath string
		if strings.Contains(config.Image, "golang") {
			// This is a golang binary
			// runCmd = "/bootstrap"
			runCmd = []string{"/bootstrap"}
			mountCmd = ":/bootstrap:ro"
			binaryPath = fmt.Sprintf("%s/binaries/%s/bootstrap", "/Users/personal/Projects/jambda", functionId)
		}

		if strings.Contains(config.Image, "jdk") {
			// This is a java binary
			runCmd = []string{"/bin/sh", "-c", "java -jar /bootstrap.jar"}
			mountCmd = ":/bootstrap.jar:ro"
			binaryPath = fmt.Sprintf("%s/binaries/%s/bootstrap.jar", "/Users/personal/Projects/jambda", functionId)
		}

		ds.log.Infof("Running command on container : '%s'", runCmd)
		ds.log.Infof("Mounting command on container : '%s'", mountCmd)
		ds.log.Infof("Binary Path is : '%s'", binaryPath)

		// Create and start the container
		cInstance, err := ds.cli.ContainerCreate(ctx, &container.Config{
			Image: config.Image,
			// TODO: Allow custom cmd params?
			Cmd: runCmd,
			Labels: map[string]string{
				"function_id": functionId,
			},
			ExposedPorts: nat.PortSet{
				nat.Port(fmt.Sprintf("%d/tcp", *config.Port)): {},
			},
		}, &container.HostConfig{
			Binds: []string{
				binaryPath + mountCmd,
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
			return "", errors.NewDockerError(fmt.Sprintf("error creating docker container: %v", err))
		}

		// Container was not found, so we created one...
		containerId = cInstance.ID

		// Start the container
		if err := ds.cli.ContainerStart(ctx, cInstance.ID, container.StartOptions{}); err != nil {
			ds.log.Error("Failed to start container: ", err)
			return "", errors.NewDockerError(fmt.Sprintf("error starting docker container: %v", err))
		}
	}

	return containerId, nil
}

func (ds *DockerService) HealthCheckContainer(ctx context.Context, containerId string, config data.FunctionConfig) error {
	timeout := time.Now().Add(30 * time.Second) // Wait up to 30 seconds

	url, err := ds.GetContainerUrl(ctx, containerId, config)
	if err != nil {
		return errors.NewDockerError(fmt.Sprintf("error inspecting container: %v", err))
	}

	for time.Now().Before(timeout) {
		resp, err := http.Get(fmt.Sprintf("%s/health", url))
		if err == nil && resp.StatusCode == http.StatusOK {
			return nil // Container is ready
		}

		// Wait a bit before checking again
		ds.log.Warn("Health check failed. Retrying.")
		time.Sleep(2 * time.Second)
	}

	return errors.NewDockerError(fmt.Sprintf("container did not become ready within the expected time"))
}

func (ds *DockerService) GetContainerUrl(ctx context.Context, containerId string, config data.FunctionConfig) (string, error) {
	inspectData, err := ds.cli.ContainerInspect(ctx, containerId)
	if err != nil {
		ds.log.Error("Error inspecting container: ", err)
		return "", err
	}

	portKey := nat.Port(fmt.Sprintf("%d/tcp", *config.Port))

	// Force safe access to port bindings
	portBindings, ok := inspectData.NetworkSettings.Ports[portKey]
	if !ok || len(portBindings) == 0 {
		ds.log.Errorf("No port bindings are available for port %s", portKey)
		return "", fmt.Errorf("no port bindings are available for port %s", portKey)
	}

	assignedPort := portBindings[0].HostPort
	if assignedPort == "" {
		ds.log.Error("Assigned host port is empty")
		return "", fmt.Errorf("assigned host port is empty for port '%s'", portKey)
	}

	return fmt.Sprintf("http://%s:%s", "localhost", assignedPort), nil
}

func (ds *DockerService) StopContainerForFunction(functionID string) {
	ds.log.Infof("Stopping idle container for function '%s' ", functionID)

	// Max wait for container stop is 10 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filterArgs := filters.NewArgs()
	filterArgs.Add("label", fmt.Sprintf("function_id=%s", functionID))
	containers, err := ds.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		ds.log.Errorf("Failed to list containers: %s", err)
	}

	ds.log.Infof("Number of containers: %d", len(containers))

	for _, inContainer := range containers {
		ds.log.Infof("Stopping Docker container for function: %s", functionID)
		opts := container.StopOptions{
			Timeout: nil,
		}
		if err := ds.cli.ContainerStop(ctx, inContainer.ID, opts); err != nil {
			ds.log.Errorf("Failed to stop container %s: %s", inContainer.ID, err)
		} else {
			ds.log.Infof("Stopped container %s for function: %s", inContainer.ID, functionID)
		}
	}
}
