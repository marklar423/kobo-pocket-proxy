package server_test

import (
	"context"
	"fmt"
	"io"
	"proxyserver/readeck"
	"strings"
	"testing"

	containers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// This file runs an integration test which spins up a real Readeck server in a container, points
// an instance of the proxy server at it, and runs Pocket-compatible API calls through the proxy
// server.
const readeckImage = "codeberg.org/readeck/readeck:0.19.2"
const readeckUser = "tester"
const readeckPassword = "testpassword"

type env struct {
	readeckContainer containers.Container
	readeckBaseUrl   string
	authToken        string
}

func (e env) cleanup(t *testing.T) {
	containers.CleanupContainer(t, e.readeckContainer)
}

func setupEnv(ctx context.Context, t *testing.T) (env, error) {
	// Create the container
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
	errCleanup := func() { containers.CleanupContainer(t, readeckContainer) }

	if err != nil {
		errCleanup()
		return env{}, err
	}
	t.Log("Readeck container started")

	e := env{
		readeckContainer: readeckContainer,
	}

	// Populate the readeck host:port
	host, err := readeckContainer.Host(ctx)
	if err != nil {
		errCleanup()
		return env{}, err
	}
	port, err := readeckContainer.MappedPort(ctx, "8000/tcp")
	if err != nil {
		errCleanup()
		return env{}, err
	}
	e.readeckBaseUrl = fmt.Sprintf("http://%s:%s", host, port.Port())
	t.Logf("Readeck base URL: %s", e.readeckBaseUrl)

	// Create the initial user
	userCmd := []string{
		"/bin/readeck", "user",
		"-config", "/readeck/config.toml",
		"-email", "test@test.com",
		"-group", "admin",
		"-p", readeckPassword,
		"-u", readeckUser,
	}
	if code, output, err := readeckContainer.Exec(ctx, userCmd); err != nil {
		errCleanup()
		return env{}, err
	} else if code != 0 {
		errCleanup()
		return env{}, fmt.Errorf("unexpected error from readeck user command: want 0 got %d", code)
	} else {
		buf := new(strings.Builder)
		if _, err := io.Copy(buf, output); err == nil {
			t.Log(buf.String())
		}
	}

	// Get the auth token.
	token, err := readeck.GetAuthToken(e.readeckBaseUrl, "test app", readeckUser, readeckPassword)
	if err != nil {
		errCleanup()
		return env{}, err
	}
	e.authToken = token
	t.Logf("Readeck auth token: %s", e.authToken)

	return e, nil
}

func TestServer(t *testing.T) {
	ctx := context.Background()

	env, err := setupEnv(ctx, t)
	if err != nil {
		t.Fatalf("Unexpected error fron setupEnv(): %v", err)
	}
	defer env.cleanup(t)
}
