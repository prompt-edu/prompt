package coreRequests

import (
	"encoding/json"
	"fmt"
	"strings"

	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/team_allocation/coreRequests/coreRequestDTO"
)

// GetCourses fetches courses from the core service.
// It returns a slice of Course objects containing course information including phases.
func GetCourses(coreURL string, authHeader string) ([]coreRequestDTO.Course, error) {
	if coreURL == "" {
		return nil, fmt.Errorf("coreURL cannot be empty")
	}

	path := "/api/courses/"
	// Ensure proper URL concatenation by trimming trailing slash from base URL
	url := strings.TrimRight(coreURL, "/") + path

	data, err := promptSDK.FetchJSON(url, authHeader)
	if err != nil {
		return nil, err
	}

	var results []coreRequestDTO.Course
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, err
	}

	return results, nil
}
