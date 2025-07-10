package server

import (
	"context"
	"testing"

	containers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// This file runs an integration test which spins up a real Readeck server in a container, points
// an instance of the proxy server at it, and runs Pocket-compatible API calls through the proxy
// server.
const readeckImage = "codeberg.org/readeck/readeck:0.19.2"

func setupEnv(ctx context.Context) (containers.Container, error) {
	req := containers.GenericContainerRequest{
		ProviderType: containers.ProviderPodman,
		ContainerRequest: containers.ContainerRequest{
			Image:        readeckImage,
			ExposedPorts: []string{"8000/tcp"},
			WaitingFor:   wait.ForExposedPort(),
		},
		Started: true,
	}

	readeckContainer, err := containers.GenericContainer(ctx, req)
	if err != nil {
		return nil, err
	}

	return readeckContainer, nil
}

func TestServer(t *testing.T) {
	ctx := context.Background()

	readeckContainer, err := setupEnv(ctx)
	if err != nil {
		t.Fatalf("Unexpected error fron setupEnv(): %v", err)
	}
	defer containers.CleanupContainer(t, readeckContainer)
}
