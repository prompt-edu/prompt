package developerProfileDTO

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/intro_course/db/sqlc"
)

type DeveloperProfile struct {
	CoursePhaseID         uuid.UUID   `json:"coursePhaseID"`
	CourseParticipationID uuid.UUID   `json:"courseParticipationID"`
	AppleID               string      `json:"appleID"`
	GitLabUsername        string      `json:"gitLabUsername"`
	HasMacBook            bool        `json:"hasMacBook"`
	IPhoneUDID            pgtype.Text `json:"iPhoneUDID"`
	IPadUDID              pgtype.Text `json:"iPadUDID"`
	AppleWatchUDID        pgtype.Text `json:"appleWatchUDID"`
}

func GetDeveloperProfileDTOFromDBModel(model db.DeveloperProfile) DeveloperProfile {
	return DeveloperProfile{
		CoursePhaseID:         model.CoursePhaseID,
		CourseParticipationID: model.CourseParticipationID,
		AppleID:               model.AppleID,
		GitLabUsername:        model.GitlabUsername,
		HasMacBook:            model.HasMacbook,
		IPhoneUDID:            model.IphoneUdid,
		IPadUDID:              model.IpadUdid,
		AppleWatchUDID:        model.AppleWatchUdid,
	}
}

func GetDeveloperProfileDTOsFromDBModels(models []db.DeveloperProfile) []DeveloperProfile {
	developerProfiles := make([]DeveloperProfile, 0, len(models))
	for _, model := range models {
		developerProfiles = append(developerProfiles, GetDeveloperProfileDTOFromDBModel(model))
	}
	return developerProfiles
}
