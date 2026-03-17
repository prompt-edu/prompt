package coreRequests

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/team_allocation/coreRequests/coreRequestDTO"
	log "github.com/sirupsen/logrus"
)

func SendAddTutorsToKeycloakGroup(authHeader string, courseID uuid.UUID, tutorIDs []uuid.UUID, groupName string) error {
	path := "/api/keycloak/" + courseID.String() + "/group/" + groupName + "/tutors"

	// Create the payload
	payload := coreRequestDTO.AddTutorsToGroup{
		Tutors: tutorIDs,
	}

	// Marshal payload to JSON
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Send the request with the payload attached
	resp, err := sendRequest("PUT", path, authHeader, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		log.Error("Received non-OK response:", resp.Status)
		return errors.New("non-OK response received")
	}

	return nil
}
