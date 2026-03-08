package developerProfileDTO

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/intro_course/db/sqlc"
)

type PostDeveloperProfile struct {
	AppleID        string      `json:"appleID"`
	GitLabUsername string      `json:"gitLabUsername"`
	HasMacBook     bool        `json:"hasMacBook"`
	IPhoneUDID     pgtype.Text `json:"iPhoneUDID"`
	IPadUDID       pgtype.Text `json:"iPadUDID"`
	AppleWatchUDID pgtype.Text `json:"appleWatchUDID"`
}

func GetDeveloperProfileDTOFromPostRequest(request PostDeveloperProfile, coursePhaseID, courseParticipationID uuid.UUID) db.CreateDeveloperProfileParams {
	return db.CreateDeveloperProfileParams{
		CoursePhaseID:         coursePhaseID,
		CourseParticipationID: courseParticipationID,
		AppleID:               request.AppleID,
		GitlabUsername:        request.GitLabUsername,
		HasMacbook:            request.HasMacBook,
		IphoneUdid:            request.IPhoneUDID,
		IpadUdid:              request.IPadUDID,
		AppleWatchUdid:        request.AppleWatchUDID,
	}
}

func GetDeveloperProfileDTOFromCreateRequest(request DeveloperProfile, coursePhaseID, courseParticipationID uuid.UUID) db.CreateOrUpdateDeveloperProfileParams {
	return db.CreateOrUpdateDeveloperProfileParams{
		CoursePhaseID:         coursePhaseID,
		CourseParticipationID: courseParticipationID,
		AppleID:               request.AppleID,
		GitlabUsername:        request.GitLabUsername,
		HasMacbook:            request.HasMacBook,
		IphoneUdid:            request.IPhoneUDID,
		IpadUdid:              request.IPadUDID,
		AppleWatchUdid:        request.AppleWatchUDID,
	}
}
