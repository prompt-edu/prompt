package participants

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	promptSDK "github.com/ls1intum/prompt-sdk"
	"github.com/ls1intum/prompt-sdk/promptTypes"
	db "github.com/ls1intum/prompt2/servers/certificate/db/sqlc"
	"github.com/ls1intum/prompt2/servers/certificate/utils"
	log "github.com/sirupsen/logrus"
)

type ParticipantsService struct {
	queries    db.Queries
	httpClient *http.Client
	coreURL    string
}

var ErrServiceNotInitialized = errors.New("participants service not initialized")

var ParticipantsServiceSingleton *ParticipantsService

func NewParticipantsService(queries db.Queries) *ParticipantsService {
	return &ParticipantsService{
		queries: queries,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		coreURL: utils.GetCoreUrl(),
	}
}

func (s *ParticipantsService) makeAuthenticatedRequest(ctx context.Context, method, url, authHeader string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return resp, nil
}

func GetParticipationsForCoursePhase(ctx context.Context, authHeader string, coursePhaseID uuid.UUID) ([]ParticipantWithDownloadStatus, error) {
	s := ParticipantsServiceSingleton
	if s == nil {
		return nil, ErrServiceNotInitialized
	}

	url := fmt.Sprintf("%s/api/course_phases/%s/participations", s.coreURL, coursePhaseID.String())
	log.WithField("url", url).Debug("Fetching participations from core service")

	resp, err := s.makeAuthenticatedRequest(ctx, "GET", url, authHeader)
	if err != nil {
		log.WithError(err).Error("Failed to fetch participations from core service")
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.WithFields(log.Fields{
			"status": resp.Status,
			"body":   string(body),
		}).Error("Failed to fetch participations from core service")
		return nil, fmt.Errorf("failed to fetch participations: %s", resp.Status)
	}

	var participationsResp CoursePhaseParticipationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&participationsResp); err != nil {
		log.WithError(err).Error("Failed to decode participations response")
		return nil, fmt.Errorf("failed to decode participations response: %w", err)
	}

	// Get download records for this course phase
	downloads, err := s.queries.ListCertificateDownloadsByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.WithError(err).Warn("Failed to fetch certificate downloads, continuing without download info")
		downloads = []db.CertificateDownload{}
	}

	// Create a map for quick lookup
	downloadMap := make(map[uuid.UUID]db.CertificateDownload)
	for _, d := range downloads {
		downloadMap[d.StudentID] = d
	}

	// Map participations with download status
	result := make([]ParticipantWithDownloadStatus, 0, len(participationsResp.CoursePhaseParticipations))
	for _, p := range participationsResp.CoursePhaseParticipations {
		participant := ParticipantWithDownloadStatus{
			CoursePhaseParticipation: p,
			HasDownloaded:            false,
			DownloadCount:            0,
		}

		if download, found := downloadMap[p.Student.ID]; found {
			participant.HasDownloaded = true
			participant.DownloadCount = download.DownloadCount
			if download.FirstDownload.Valid {
				t := download.FirstDownload.Time
				participant.FirstDownload = &t
			}
			if download.LastDownload.Valid {
				t := download.LastDownload.Time
				participant.LastDownload = &t
			}
		}

		result = append(result, participant)
	}

	return result, nil
}

func GetCoursePhaseWithCourse(ctx context.Context, authHeader string, coursePhaseID uuid.UUID) (*CoursePhaseWithCourse, error) {
	s := ParticipantsServiceSingleton
	if s == nil {
		return nil, ErrServiceNotInitialized
	}

	url := fmt.Sprintf("%s/api/course_phases/%s", s.coreURL, coursePhaseID.String())
	log.WithField("url", url).Debug("Fetching course phase info from core service")

	resp, err := s.makeAuthenticatedRequest(ctx, "GET", url, authHeader)
	if err != nil {
		log.WithError(err).Error("Failed to fetch course phase info from core service")
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.WithFields(log.Fields{
			"status": resp.Status,
			"body":   string(body),
		}).Error("Failed to fetch course phase info from core service")
		return nil, fmt.Errorf("failed to fetch course phase info: %s", resp.Status)
	}

	var coursePhase CoursePhaseWithCourse
	if err := json.NewDecoder(resp.Body).Decode(&coursePhase); err != nil {
		log.WithError(err).Error("Failed to decode course phase response")
		return nil, fmt.Errorf("failed to decode course phase response: %w", err)
	}

	return &coursePhase, nil
}

func GetStudentInfo(ctx context.Context, authHeader string, coursePhaseID, studentID uuid.UUID) (*Student, error) {
	s := ParticipantsServiceSingleton
	if s == nil {
		return nil, ErrServiceNotInitialized
	}

	// Use the /participations/students endpoint which returns students by their core student ID
	url := fmt.Sprintf("%s/api/course_phases/%s/participations/students", s.coreURL, coursePhaseID.String())
	log.WithField("url", url).Debug("Fetching students from core service")

	resp, err := s.makeAuthenticatedRequest(ctx, "GET", url, authHeader)
	if err != nil {
		log.WithError(err).Error("Failed to fetch students from core service")
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.WithFields(log.Fields{
			"status": resp.Status,
			"body":   string(body),
		}).Error("Failed to fetch students from core service")
		return nil, fmt.Errorf("failed to fetch students: %s", resp.Status)
	}

	var students []Student
	if err := json.NewDecoder(resp.Body).Decode(&students); err != nil {
		log.WithError(err).Error("Failed to decode students response")
		return nil, fmt.Errorf("failed to decode students response: %w", err)
	}

	for _, student := range students {
		if student.ID == studentID {
			return &student, nil
		}
	}

	return nil, fmt.Errorf("student %s not found in course phase %s", studentID, coursePhaseID)
}

// GetOwnStudentInfo fetches the current student's info from the core's /self endpoint.
// This returns the core student ID (not the Keycloak UUID), which is needed for
// correctly recording certificate downloads.
func GetOwnStudentInfo(ctx context.Context, authHeader string, coursePhaseID uuid.UUID) (*Student, error) {
	s := ParticipantsServiceSingleton
	if s == nil {
		return nil, ErrServiceNotInitialized
	}

	url := fmt.Sprintf("%s/api/course_phases/%s/participations/self", s.coreURL, coursePhaseID.String())
	log.WithField("url", url).Debug("Fetching own student info from core service")

	resp, err := s.makeAuthenticatedRequest(ctx, "GET", url, authHeader)
	if err != nil {
		log.WithError(err).Error("Failed to fetch own student info from core service")
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.WithFields(log.Fields{
			"status": resp.Status,
			"body":   string(body),
		}).Error("Failed to fetch own student info from core service")
		return nil, fmt.Errorf("failed to fetch own student info: %s", resp.Status)
	}

	var participation CoursePhaseParticipationSelf
	if err := json.NewDecoder(resp.Body).Decode(&participation); err != nil {
		log.WithError(err).Error("Failed to decode own student response")
		return nil, fmt.Errorf("failed to decode own student response: %w", err)
	}

	return &participation.Student, nil
}

// GetStudentTeamName resolves the team name for a student in a given course phase.
// It uses the prompt-sdk resolution system to fetch:
// 1. The list of teams (with names) from the course phase data resolutions
// 2. The student's team allocation UUID from participation resolutions
// Then matches them to return the team name.
// Returns empty string (no error) if no team is allocated.
func GetStudentTeamName(ctx context.Context, authHeader string, coursePhaseID, studentID uuid.UUID) (string, error) {
	s := ParticipantsServiceSingleton
	if s == nil {
		return "", ErrServiceNotInitialized
	}
	coreURL := s.coreURL

	// Fetch teams from course phase data resolutions
	cpData, err := promptSDK.FetchAndMergeCoursePhaseWithResolution(coreURL, authHeader, coursePhaseID)
	if err != nil {
		log.WithError(err).Warn("Could not fetch course phase resolutions for teams")
		return "", nil
	}

	teamsRaw, teamsExist := cpData["teams"]
	if !teamsExist {
		log.Debug("No teams found in course phase data")
		return "", nil
	}

	teams, err := parseTeams(teamsRaw)
	if err != nil {
		log.WithError(err).Warn("Failed to parse teams from course phase data")
		return "", nil
	}

	if len(teams) == 0 {
		return "", nil
	}

	// Fetch all participations with resolutions to find the student's team allocation
	participations, err := promptSDK.FetchAndMergeParticipationsWithResolutions(coreURL, authHeader, coursePhaseID)
	if err != nil {
		log.WithError(err).Warn("Could not fetch participations with resolutions for team lookup")
		return "", nil
	}

	// Find the participation matching this student
	var teamIDStr string
	for _, p := range participations {
		if p.Student.ID == studentID {
			if val, ok := p.PrevData["teamAllocation"].(string); ok {
				teamIDStr = val
			}
			break
		}
	}

	if teamIDStr == "" {
		log.Debug("No team allocation found for student")
		return "", nil
	}

	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		log.WithError(err).Warn("Invalid team allocation UUID")
		return "", nil
	}

	// Find team name by ID
	for _, team := range teams {
		if team.ID == teamID {
			return team.Name, nil
		}
	}

	log.WithField("teamID", teamID).Warn("Team ID from allocation not found in teams list")
	return "", nil
}

// parseTeams parses the teams slice from the course phase resolution metadata.
func parseTeams(teamsRaw interface{}) ([]promptTypes.Team, error) {
	teams := make([]promptTypes.Team, 0)
	if teamsRaw == nil {
		return teams, nil
	}

	teamsSlice, ok := teamsRaw.([]interface{})
	if !ok {
		return nil, errors.New("invalid teams data structure")
	}

	for i, teamData := range teamsSlice {
		if team, ok := parseTeam(teamData, i); ok {
			teams = append(teams, team)
		}
	}
	return teams, nil
}

// parseTeam parses individual team data from a map interface.
func parseTeam(teamData interface{}, index int) (promptTypes.Team, bool) {
	teamMap, ok := teamData.(map[string]interface{})
	if !ok {
		log.Warnf("Skipping team at index %d: not a valid map", index)
		return promptTypes.Team{}, false
	}

	teamIDStr, ok := teamMap["id"].(string)
	if !ok {
		log.Warnf("Skipping team at index %d: missing or invalid 'id'", index)
		return promptTypes.Team{}, false
	}
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		log.Warnf("Skipping team at index %d: invalid UUID: %v", index, err)
		return promptTypes.Team{}, false
	}

	teamName, ok := teamMap["name"].(string)
	if !ok {
		log.Warnf("Skipping team at index %d: missing or invalid 'name'", index)
		return promptTypes.Team{}, false
	}

	members := parsePersons(teamMap["members"])
	tutors := parsePersons(teamMap["tutors"])

	return promptTypes.Team{
		ID:      teamID,
		Name:    teamName,
		Members: members,
		Tutors:  tutors,
	}, true
}

// parsePersons parses a slice of person data from a raw interface.
func parsePersons(personsRaw interface{}) []promptTypes.Person {
	persons := make([]promptTypes.Person, 0)
	if personsRaw == nil {
		return persons
	}

	personsSlice, ok := personsRaw.([]interface{})
	if !ok {
		return persons
	}

	for _, personData := range personsSlice {
		memberMap, ok := personData.(map[string]interface{})
		if !ok {
			continue
		}
		id, err := uuid.Parse(fmt.Sprintf("%v", memberMap["id"]))
		if err != nil {
			continue
		}
		firstName, _ := memberMap["firstName"].(string)
		lastName, _ := memberMap["lastName"].(string)
		persons = append(persons, promptTypes.Person{
			ID:        id,
			FirstName: firstName,
			LastName:  lastName,
		})
	}
	return persons
}
