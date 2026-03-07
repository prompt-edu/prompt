package infrastructureDTO

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/intro_course/db/sqlc"
)

type GitlabStatus struct {
	CoursePhaseID         uuid.UUID        `json:"coursePhaseID"`
	CourseParticipationID uuid.UUID        `json:"courseParticipationID"`
	GitlabSuccess         bool             `json:"gitlabSuccess"`
	ErrorMessage          string           `json:"errorMessage"`
	CreatedAt             pgtype.Timestamp `json:"createdAt"`
	UpdatedAt             pgtype.Timestamp `json:"updatedAt"`
}

func getGitlabStatusDTOFromModel(model db.StudentGitlabProcess) GitlabStatus {
	return GitlabStatus{
		CoursePhaseID:         model.CoursePhaseID,
		CourseParticipationID: model.CourseParticipationID,
		GitlabSuccess:         model.GitlabSuccess,
		ErrorMessage:          model.ErrorMessage.String,
		CreatedAt:             model.CreatedAt,
		UpdatedAt:             model.UpdatedAt,
	}
}

func GetGitlabStatusDTOsFromModels(models []db.StudentGitlabProcess) []GitlabStatus {
	dtos := make([]GitlabStatus, 0, len(models))
	for _, model := range models {
		dtos = append(dtos, getGitlabStatusDTOFromModel(model))
	}
	return dtos
}
