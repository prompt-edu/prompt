package coursePhaseConfigDTO

import (
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
)

type AssessmentParticipationWithStudent struct {
	promptTypes.CoursePhaseParticipationWithStudent
	TeamID *uuid.UUID `json:"teamID,omitempty"`
}

func GetAssessmentStudentsFromParticipations(participations []promptTypes.CoursePhaseParticipationWithStudent) []AssessmentParticipationWithStudent {
	assessmentStudents := make([]AssessmentParticipationWithStudent, len(participations))

	for i, participation := range participations {
		assessmentStudents[i] = GetAssessmentStudentFromParticipation(participation)
	}

	return assessmentStudents
}

func GetAssessmentStudentFromParticipation(participation promptTypes.CoursePhaseParticipationWithStudent) AssessmentParticipationWithStudent {
	teamIDStr, ok := participation.PrevData["teamAllocation"].(string)
	var teamID *uuid.UUID
	if ok {
		parsedTeamID, err := uuid.Parse(teamIDStr)
		if err == nil {
			teamID = &parsedTeamID
		}
	}

	return AssessmentParticipationWithStudent{
		CoursePhaseParticipationWithStudent: participation,
		TeamID:                              teamID,
	}
}
