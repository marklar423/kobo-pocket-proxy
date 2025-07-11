package server_test

import (
	"context"
	"fmt"
	"testing"

	containers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// This file runs an integration test which spins up a real Readeck server in a container, points
// an instance of the proxy server at it, and runs Pocket-compatible API calls through the proxy
// server.
const readeckImage = "codeberg.org/readeck/readeck:0.19.2"

type env struct {
	readeckContainer containers.Container
}

func (e env) cleanup(t *testing.T) {
	containers.CleanupContainer(t, e.readeckContainer)
}

func setupEnv(ctx context.Context) (env, error) {
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
		return env{}, err
	}

	// Create the initial user
	userCmd := []string{
		"/bin/readeck", "user",
		"-config", "/readeck/config.toml",
		"-email", "test@test.com",
		"-group", "admin",
		"-p", "test",
		"-u", "tester",
	}
	if code, _, err := readeckContainer.Exec(ctx, userCmd); err != nil {
		return env{}, err
	} else if code != 0 {
		return env{}, fmt.Errorf("unexpected error from readeck user command: want 0 got %d", code)
	}

	return env{
		readeckContainer: readeckContainer,
	}, nil
}

func TestServer(t *testing.T) {
	ctx := context.Background()

	env, err := setupEnv(ctx)
	if err != nil {
		t.Fatalf("Unexpected error fron setupEnv(): %v", err)
	}
	defer env.cleanup(t)
}
