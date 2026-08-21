package coreRequests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/team_allocation/coreRequests/coreRequestDTO"
	log "github.com/sirupsen/logrus"
)

func SendAddTutorsToKeycloakGroup(ctx context.Context, authHeader string, courseID uuid.UUID, tutorIDs []uuid.UUID, groupName string) error {
	url := sdkUtils.GetCoreUrl() + "/api/keycloak/" + courseID.String() + "/group/" + groupName + "/tutors"

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
	resp, err := sdkUtils.SendCoreRequest(ctx, http.MethodPut, authHeader, bytes.NewBuffer(body), url)
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
