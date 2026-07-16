package presentation

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
)

type ownParticipationWithResolutions struct {
	CourseParticipationID uuid.UUID              `json:"courseParticipationID"`
	Student               promptTypes.Student    `json:"student"`
	Resolutions           []promptSDK.Resolution `json:"resolutions"`
}

type teamTarget struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func authorizationHeader(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return value
}

func (s *Service) fetchIndividualTargets(authHeader string, coursePhaseID uuid.UUID) ([]TargetResponse, error) {
	participations, err := promptSDK.FetchAndMergeParticipationsWithResolutions(s.coreURL, authHeader, coursePhaseID)
	if err != nil {
		return nil, fmt.Errorf("fetch course phase participations: %w", err)
	}

	targets := make([]TargetResponse, 0, len(participations))
	for _, participation := range participations {
		name := strings.TrimSpace(participation.Student.FirstName + " " + participation.Student.LastName)
		targets = append(targets, TargetResponse{
			ID:   participation.CourseParticipationID,
			Type: "individual",
			Name: name,
		})
	}
	return targets, nil
}

func (s *Service) fetchTeamTargets(authHeader string, coursePhaseID uuid.UUID) ([]TargetResponse, error) {
	phaseData, err := promptSDK.FetchAndMergeCoursePhaseWithResolution(s.coreURL, authHeader, coursePhaseID)
	if err != nil {
		return nil, fmt.Errorf("fetch course phase team data: %w", err)
	}
	rawTeams, exists := phaseData["teams"]
	if !exists || rawTeams == nil {
		return []TargetResponse{}, nil
	}
	data, err := json.Marshal(rawTeams)
	if err != nil {
		return nil, fmt.Errorf("marshal resolved teams: %w", err)
	}
	var teams []teamTarget
	if err := json.Unmarshal(data, &teams); err != nil {
		return nil, fmt.Errorf("decode resolved teams: %w", err)
	}
	targets := make([]TargetResponse, 0, len(teams))
	for _, team := range teams {
		targets = append(targets, TargetResponse{ID: team.ID, Type: "team", Name: team.Name})
	}
	return targets, nil
}

func (s *Service) fetchOwnTeamID(authHeader string, coursePhaseID, courseParticipationID uuid.UUID) (uuid.UUID, error) {
	url := strings.TrimRight(s.coreURL, "/") + "/api/course_phases/" + coursePhaseID.String() + "/participations/self"
	data, err := promptSDK.FetchJSON(url, authHeader)
	if err != nil {
		return uuid.Nil, fmt.Errorf("fetch own course phase participation: %w", err)
	}
	var own ownParticipationWithResolutions
	if err := json.Unmarshal(data, &own); err != nil {
		return uuid.Nil, fmt.Errorf("decode own course phase participation: %w", err)
	}
	if own.CourseParticipationID != uuid.Nil {
		courseParticipationID = own.CourseParticipationID
	}
	for _, resolution := range own.Resolutions {
		if resolution.DtoName != "teamAllocation" {
			continue
		}
		resolved, err := promptSDK.ResolveParticipation(authHeader, resolution, courseParticipationID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("resolve current team allocation: %w", err)
		}
		switch value := resolved.(type) {
		case string:
			teamID, parseErr := uuid.Parse(value)
			if parseErr != nil {
				return uuid.Nil, fmt.Errorf("parse resolved team allocation: %w", parseErr)
			}
			return teamID, nil
		case map[string]interface{}:
			raw, ok := value["teamAllocation"].(string)
			if !ok {
				break
			}
			teamID, parseErr := uuid.Parse(raw)
			if parseErr != nil {
				return uuid.Nil, fmt.Errorf("parse resolved team allocation: %w", parseErr)
			}
			return teamID, nil
		}
	}
	return uuid.Nil, apiError(403, "team_not_resolved", "No current team allocation is available", nil)
}
