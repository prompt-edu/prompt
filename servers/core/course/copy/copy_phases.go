package copy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt/servers/core/coursePhase"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseDTO"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/resolution"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	log "github.com/sirupsen/logrus"
)

// copyCoursePhases duplicates all phases from the source course to the target course.
// It returns a mapping of old phase IDs to new phase IDs.
func copyCoursePhases(c *gin.Context, qtx *db.Queries, sourceID, targetID uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	sequence, err := qtx.GetCoursePhaseSequence(c, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get course phase sequence: %w", err)
	}
	unordered, err := qtx.GetNotOrderedCoursePhases(c, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unordered course phases: %w", err)
	}

	allPhases := make(map[uuid.UUID]struct{})
	for _, p := range sequence {
		allPhases[p.ID] = struct{}{}
	}
	for _, p := range unordered {
		allPhases[p.ID] = struct{}{}
	}

	mapping := make(map[uuid.UUID]uuid.UUID)
	for oldID := range allPhases {
		phase, err := coursePhase.GetCoursePhaseByID(c, oldID)
		if err != nil {
			return nil, fmt.Errorf("failed to get course phase by ID %s: %w", oldID, err)
		}

		sanitizedRestrictedData := make(map[string]interface{}, len(phase.RestrictedData))
		for k, v := range phase.RestrictedData {
			sanitizedRestrictedData[k] = v
		}
		delete(sanitizedRestrictedData, "applicationStartDate")
		delete(sanitizedRestrictedData, "applicationEndDate")

		newPhase := coursePhaseDTO.CreateCoursePhase{
			Name:                phase.Name,
			IsInitialPhase:      phase.IsInitialPhase,
			CourseID:            targetID,
			CoursePhaseTypeID:   phase.CoursePhaseTypeID,
			RestrictedData:      sanitizedRestrictedData,
			StudentReadableData: phase.StudentReadableData,
		}
		dbModel, err := newPhase.GetDBModel()
		if err != nil {
			return nil, fmt.Errorf("failed to convert phase to DB model: %w", err)
		}
		dbModel.ID = uuid.New()
		if _, err := qtx.CreateCoursePhase(c, dbModel); err != nil {
			return nil, fmt.Errorf("failed to create course phase: %w", err)
		}
		mapping[oldID] = dbModel.ID
	}
	return mapping, nil
}

// setInitialPhase sets the initial course phase in the target course by mapping
// the initial phase from the source course via the provided phase ID mapping.
func setInitialPhase(c *gin.Context, qtx *db.Queries, sourceID, targetID uuid.UUID, phaseMap map[uuid.UUID]uuid.UUID) error {
	sequence, err := qtx.GetCoursePhaseSequence(c, sourceID)
	if err != nil {
		return fmt.Errorf("failed to get phase sequence: %w", err)
	}
	for _, p := range sequence {
		if p.IsInitialPhase {
			if err := qtx.UpdateInitialCoursePhase(c, db.UpdateInitialCoursePhaseParams{
				CourseID: targetID,
				ID:       phaseMap[p.ID],
			}); err != nil {
				return fmt.Errorf("failed to set initial phase: %w", err)
			}
			break
		}
	}
	return nil
}

// copyPhaseConfigurations sends a request to the phase service to copy configurations
// for each phase that has a server-side implementation in the source course. It uses the phase ID mapping to ensure
// the correct phases are targeted.
func copyPhaseConfigurations(c *gin.Context, phaseIDMap map[uuid.UUID]uuid.UUID) error {
	for oldPhaseID, newPhaseID := range phaseIDMap {
		oldPhase, err := coursePhase.GetCoursePhaseByID(c, oldPhaseID)
		if err != nil {
			return fmt.Errorf("course phase with ID %s not found: %w", oldPhaseID, err)
		}

		oldPhaseType, err := CourseCopyServiceSingleton.queries.GetCoursePhaseTypeByID(c, oldPhase.CoursePhaseTypeID)
		if err != nil {
			return fmt.Errorf("failed to fetch course phase type: %w", err)
		}

		if oldPhaseType.BaseUrl == "core" {
			continue
		}

		coreHost := resolution.NormaliseHost(promptSDK.GetEnv("CORE_HOST", "http://localhost:8080"))
		log.Infof("Core host is '%s'", coreHost)
		baseURL := strings.ReplaceAll(oldPhaseType.BaseUrl, "{CORE_HOST}", coreHost)

		url, err := url.JoinPath(baseURL, "copy")
		if err != nil {
			return fmt.Errorf("failed to join copy path: %w", err)
		}

		body, _ := json.Marshal(promptTypes.PhaseCopyRequest{
			SourceCoursePhaseID: oldPhaseID,
			TargetCoursePhaseID: newPhaseID,
		})

		resp, err := sendRequest("POST", c.GetHeader("Authorization"), bytes.NewBuffer(body), url)
		log.Infof("Sending copy request to %s for phase %s", url, oldPhase.Name)
		if err != nil {
			return fmt.Errorf("failed to send copy request to phase service: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			if resp.StatusCode == http.StatusNotFound {
				log.Warnf("Copy functionality not found for phase service '%s', skipping copy for this phase.", oldPhaseType.Name)
				continue
			}
			return fmt.Errorf("received non-OK response from phase service '%s': %s", oldPhaseType.Name, resp.Status)
		}
	}
	return nil
}

func sendRequest(method, authHeader string, body io.Reader, url string) (*http.Response, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Error("Error creating request:", err)
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error sending request:", err)
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}

	return resp, nil
}
