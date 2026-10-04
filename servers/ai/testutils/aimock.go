package testutils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Keep in sync with docker-compose.e2e.yml.
const AIMockImage = "ghcr.io/copilotkit/aimock:1.43.0"

const TestModel = "prompt-test-model"

type AIMock struct {
	BaseURL string
}

func StartAIMock(ctx context.Context) (*AIMock, func(), error) {
	_, thisFile, _, _ := runtime.Caller(0)
	fixtures, err := filepath.Glob(filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "e2e", "fixtures", "aimock", "*.json"))
	if err != nil || len(fixtures) == 0 {
		return nil, nil, fmt.Errorf("find aimock fixtures: %v", err)
	}
	files := make([]testcontainers.ContainerFile, 0, len(fixtures))
	for _, fixture := range fixtures {
		files = append(files, testcontainers.ContainerFile{
			HostFilePath:      fixture,
			ContainerFilePath: "/fixtures/" + filepath.Base(fixture),
			FileMode:          0o644,
		})
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        AIMockImage,
			Cmd:          []string{"--fixtures", "/fixtures", "--host", "0.0.0.0"},
			ExposedPorts: []string{"4010/tcp"},
			Files:        files,
			WaitingFor:   wait.ForHTTP("/health").WithPort("4010/tcp"),
		},
		Started: true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("start aimock: %w", err)
	}
	terminate := func() { _ = container.Terminate(ctx) }
	endpoint, err := container.PortEndpoint(ctx, "4010/tcp", "http")
	if err != nil {
		terminate()
		return nil, nil, fmt.Errorf("get aimock endpoint: %w", err)
	}
	return &AIMock{BaseURL: endpoint}, terminate, nil
}

type JournalEntry struct {
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
	Body    map[string]any    `json:"body"`
}

func (m *AIMock) Journal() ([]JournalEntry, error) {
	response, err := http.Get(m.BaseURL + "/__aimock/journal")
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()
	var entries []JournalEntry
	return entries, json.NewDecoder(response.Body).Decode(&entries)
}

func (m *AIMock) ResetJournal() error {
	response, err := http.Post(m.BaseURL+"/__aimock/reset/journal", "application/json", nil)
	if err != nil {
		return err
	}
	return response.Body.Close()
}
