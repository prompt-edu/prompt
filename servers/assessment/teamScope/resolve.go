package teamScope

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
)

const teamsDtoName = "teams"

var errTeamsForbidden = errors.New("team phase denied the caller every team")

var httpClient = &http.Client{Timeout: 10 * time.Second}

type teamsResponse struct {
	Teams []struct {
		Members []struct {
			ID uuid.UUID `json:"id"`
		} `json:"members"`
	} `json:"teams"`
}

// resolveEditorScope derives an editor's scope from what the phase providing the teams
// returns for the caller's own token, so assessment keeps no tutor data of its own.
func resolveEditorScope(authHeader string, coursePhaseID uuid.UUID) (Scope, error) {
	resolution, found, err := findTeamsResolution(authHeader, coursePhaseID)
	if err != nil {
		return Scope{}, err
	}
	if !found {
		return Scope{}, nil
	}

	teams, err := fetchTeams(authHeader, resolution)
	if errors.Is(err, errTeamsForbidden) {
		return ConfinedTo(nil), nil
	}
	if err != nil {
		return Scope{}, err
	}

	var members []uuid.UUID
	for _, team := range teams.Teams {
		for _, member := range team.Members {
			members = append(members, member.ID)
		}
	}
	return ConfinedTo(members), nil
}

func findTeamsResolution(authHeader string, coursePhaseID uuid.UUID) (promptSDK.Resolution, bool, error) {
	coursePhaseDataURL, err := url.JoinPath(sdkUtils.GetCoreUrl(), "api/course_phases", coursePhaseID.String(), "course_phase_data")
	if err != nil {
		return promptSDK.Resolution{}, false, err
	}
	body, err := promptSDK.FetchJSON(coursePhaseDataURL, authHeader)
	if err != nil {
		return promptSDK.Resolution{}, false, fmt.Errorf("could not fetch course phase data: %w", err)
	}

	var coursePhaseData promptSDK.PrevCoursePhaseData
	if err := json.Unmarshal(body, &coursePhaseData); err != nil {
		return promptSDK.Resolution{}, false, fmt.Errorf("could not parse course phase data: %w", err)
	}

	for _, resolution := range coursePhaseData.Resolutions {
		if resolution.DtoName == teamsDtoName {
			return resolution, true, nil
		}
	}
	return promptSDK.Resolution{}, false, nil
}

func fetchTeams(authHeader string, resolution promptSDK.Resolution) (teamsResponse, error) {
	teamsURL, err := resolutionURL(resolution)
	if err != nil {
		return teamsResponse{}, err
	}

	status, body, err := getAsCaller(teamsURL, authHeader)
	if err != nil {
		return teamsResponse{}, fmt.Errorf("could not fetch teams: %w", err)
	}
	switch status {
	case http.StatusOK:
	case http.StatusForbidden:
		return teamsResponse{}, errTeamsForbidden
	default:
		return teamsResponse{}, fmt.Errorf("could not fetch teams: received status %d", status)
	}

	var teams teamsResponse
	if err := json.Unmarshal(body, &teams); err != nil {
		return teamsResponse{}, fmt.Errorf("could not parse teams: %w", err)
	}
	return teams, nil
}

// ResolveOwnTutorID returns the tutor id the caller has in the phase providing the teams, so
// assessment can match them to the tutor evaluations written about them. found is false when the
// phase has no teams input or the team phase does not list the caller as a tutor.
func ResolveOwnTutorID(authHeader string, coursePhaseID uuid.UUID) (tutorID uuid.UUID, found bool, err error) {
	resolution, found, err := findTeamsResolution(authHeader, coursePhaseID)
	if err != nil || !found {
		return uuid.Nil, false, err
	}

	tutorURL, err := resolutionURL(resolution, "tutors", "me")
	if err != nil {
		return uuid.Nil, false, err
	}

	status, body, err := getAsCaller(tutorURL, authHeader)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("could not fetch the caller's tutor record: %w", err)
	}
	switch status {
	case http.StatusOK:
	// A team phase without the endpoint answers 404 as well, so it reads as "not a tutor".
	case http.StatusNotFound, http.StatusForbidden:
		return uuid.Nil, false, nil
	default:
		return uuid.Nil, false, fmt.Errorf("could not fetch the caller's tutor record: received status %d", status)
	}

	var tutor promptTypes.Person
	if err := json.Unmarshal(body, &tutor); err != nil {
		return uuid.Nil, false, fmt.Errorf("could not parse the caller's tutor record: %w", err)
	}
	if tutor.ID == uuid.Nil {
		return uuid.Nil, false, nil
	}
	return tutor.ID, true, nil
}

// getAsCaller forwards the caller's own token, so the other phase applies the caller's
// permissions. The body is only read for a 200.
func getAsCaller(targetURL, authHeader string) (int, []byte, error) {
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Authorization", authHeader)

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil, nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, fmt.Errorf("could not read response: %w", err)
	}
	return resp.StatusCode, body, nil
}

// resolutionURL mirrors the SDK's unexported URL check: the caller's token must only reach
// a plain http(s) host on the resolution's own path, optionally extended by fixed segments.
func resolutionURL(resolution promptSDK.Resolution, extraSegments ...string) (string, error) {
	base, err := url.Parse(resolution.BaseURL)
	if err != nil {
		return "", fmt.Errorf("invalid resolution baseURL: %w", err)
	}
	if (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" || base.User != nil {
		return "", fmt.Errorf("resolution baseURL %q is not an allowed target", resolution.BaseURL)
	}

	endpointPath, err := url.PathUnescape(strings.Trim(resolution.EndpointPath, "/"))
	if err != nil || slices.Contains(strings.Split(endpointPath, "/"), "..") {
		return "", fmt.Errorf("resolution endpoint path %q is not allowed", resolution.EndpointPath)
	}

	segments := append([]string{"course_phase", resolution.CoursePhaseID.String(), endpointPath}, extraSegments...)
	return base.JoinPath(segments...).String(), nil
}
