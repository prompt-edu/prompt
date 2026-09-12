package coreRequests

import (
	"encoding/json"
	"net/url"

	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
)

// Participant is the subset of a course phase participation the allocation
// module needs: proof of membership plus the names cached on an allocation.
type Participant struct {
	FirstName string
	LastName  string
}

// participationsResponse mirrors only the fields we read from core's
// participations payload. The resolutions it also returns are deliberately not
// followed, so an unrelated phase module cannot fail an allocation write.
type participationsResponse struct {
	Participations []struct {
		CourseParticipationID uuid.UUID `json:"courseParticipationID"`
		Student               struct {
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		} `json:"student"`
	} `json:"participations"`
}

// GetCoursePhaseParticipants fetches every participant of a course phase from
// the core service, keyed by course participation ID.
func GetCoursePhaseParticipants(coreURL, authHeader string, coursePhaseID uuid.UUID) (map[uuid.UUID]Participant, error) {
	requestURL, err := url.JoinPath(coreURL, "api/course_phases", coursePhaseID.String(), "participations")
	if err != nil {
		return nil, err
	}

	data, err := promptSDK.FetchJSON(requestURL, authHeader)
	if err != nil {
		return nil, err
	}

	var response participationsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	participants := make(map[uuid.UUID]Participant, len(response.Participations))
	for _, participation := range response.Participations {
		participants[participation.CourseParticipationID] = Participant{
			FirstName: participation.Student.FirstName,
			LastName:  participation.Student.LastName,
		}
	}
	return participants, nil
}
