package copy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/resolution"
	log "github.com/sirupsen/logrus"
)

// checkAllCoursePhasesCopyable checks if all course phases from the source course can be copied to the target course.
func checkAllCoursePhasesCopyable(c *gin.Context, sourceCourseID uuid.UUID) ([]string, error) {
	sequence, err := CourseCopyServiceSingleton.queries.GetCoursePhaseSequence(c, sourceCourseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get course phase sequence: %w", err)
	}
	unordered, err := CourseCopyServiceSingleton.queries.GetNotOrderedCoursePhases(c, sourceCourseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unordered course phases: %w", err)
	}

	checkedCoursePhases := make(map[uuid.UUID]bool)
	missingPhases := []string{}

	for _, p := range sequence {
		if err := checkPhaseCopyable(c, p.ID, p.CoursePhaseTypeID, p.Name.String, checkedCoursePhases, &missingPhases); err != nil {
			return nil, fmt.Errorf("failed to check phase copyable: %w", err)
		}
	}

	for _, p := range unordered {
		if err := checkPhaseCopyable(c, p.ID, p.CoursePhaseTypeID, p.Name.String, checkedCoursePhases, &missingPhases); err != nil {
			return nil, fmt.Errorf("failed to check phase copyable: %w", err)
		}
	}

	return missingPhases, nil
}

// checkPhaseCopyable checks if a single course phase can be copied by sending a dummy request to the copy endpoint.
func checkPhaseCopyable(c *gin.Context, phaseID, phaseTypeID uuid.UUID, phaseName string, checkedCoursePhases map[uuid.UUID]bool, missingPhases *[]string) error {
	coursePhaseType, err := CourseCopyServiceSingleton.queries.GetCoursePhaseTypeByID(c, phaseTypeID)
	if err != nil {
		return fmt.Errorf("failed to get phase type: %w", err)
	}

	if coursePhaseType.BaseUrl == "core" {
		return nil
	}

	coreHost := resolution.NormaliseHost(promptSDK.GetEnv("CORE_HOST", "http://localhost:8080"))
	log.Infof("Core host is '%s'", coreHost)
	baseURL := strings.ReplaceAll(coursePhaseType.BaseUrl, "{CORE_HOST}", coreHost)

	url, err := url.JoinPath(baseURL, "copy")
	if err != nil {
		return fmt.Errorf("failed to join copy path: %w", err)
	}

	// send a dummy POST request to the copy endpoint to check if it exists
	body, _ := json.Marshal(promptTypes.PhaseCopyRequest{
		SourceCoursePhaseID: phaseID,
		TargetCoursePhaseID: phaseID,
	})

	resp, err := sendRequest("POST", c.GetHeader("Authorization"), bytes.NewBuffer(body), url)
	log.Infof("Checking copy endpoint for phase '%s' with url %s", coursePhaseType.Name, url)
	if err != nil {
		log.Warnf("Error checking copy endpoint for phase '%s': %v", coursePhaseType.Name, err)
		*missingPhases = append(*missingPhases, phaseName+" ("+coursePhaseType.Name+")")
		checkedCoursePhases[phaseTypeID] = true
		return nil
	}
	_ = resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		*missingPhases = append(*missingPhases, phaseName+" ("+coursePhaseType.Name+")")
	}
	checkedCoursePhases[phaseTypeID] = true
	return nil
}
