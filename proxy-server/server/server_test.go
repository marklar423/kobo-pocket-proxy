package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
	readeckBaseUrl   string
	authToken        string
}

func (e env) cleanup(t *testing.T) {
	containers.CleanupContainer(t, e.readeckContainer)
}

func getAuthToken(baseUrl string) (string, error) {
	url := fmt.Sprintf("%s/api/auth", baseUrl)
	payload := []byte(`{
		"application": "test",
		"password": "test",
		"username": "tester",
	}`)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error calling Readeck auth API: [%d] %s", res.StatusCode, res.Status)
	}

	var resBody struct {
		Token string
	}
	if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		return "", err
	}
	if resBody.Token == "" {
		return "", errors.New("unexpected empty token in Readeck auth response")
	}

	return resBody.Token, nil
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
	if err != nil {
		return env{}, err
	}
	t.Log("Readeck container started")

	e := env{
		readeckContainer: readeckContainer,
	}

	// Populate the readeck host:port
	host, err := readeckContainer.Host(ctx)
	if err != nil {
		return env{}, err
	}
	port, err := readeckContainer.MappedPort(ctx, "8000/tcp")
	if err != nil {
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
		"-p", "test",
		"-u", "tester",
	}
	if code, _, err := readeckContainer.Exec(ctx, userCmd); err != nil {
		return env{}, err
	} else if code != 0 {
		return env{}, fmt.Errorf("unexpected error from readeck user command: want 0 got %d", code)
	}

	// Get the auth token.
	token, err := getAuthToken(e.readeckBaseUrl)
	if err != nil {
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
