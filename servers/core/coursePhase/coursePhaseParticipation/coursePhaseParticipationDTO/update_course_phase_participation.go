package coursePhaseParticipationDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/meta"
	log "github.com/sirupsen/logrus"
)

type UpdateCoursePhaseParticipationRequest struct {
	// for individual updates, the courseParticipation is in the url
	// for batch updates, the ID is in the body
	CoursePhaseID         uuid.UUID      `json:"coursePhaseID"`
	CourseParticipationID uuid.UUID      `json:"courseParticipationID"`
	PassStatus            *db.PassStatus `json:"passStatus"`
	RestrictedData        meta.MetaData  `json:"restrictedData"`
	StudentReadableData   meta.MetaData  `json:"studentReadableData"`
}

type UpdateCoursePhaseParticipation struct {
	CoursePhaseID         uuid.UUID      `json:"coursePhaseID"`
	CourseParticipationID uuid.UUID      `json:"courseParticipationID"`
	PassStatus            *db.PassStatus `json:"passStatus"`
	RestrictedData        meta.MetaData  `json:"restrictedData"`
	StudentReadableData   meta.MetaData  `json:"studentReadableData"`
}

func (c UpdateCoursePhaseParticipation) GetDBModel() (db.UpdateCoursePhaseParticipationParams, error) {
	restrictedData, err := c.RestrictedData.GetDBModel()
	if err != nil {
		log.Error("failed to create CoursePhaseParticipation DB model from DTO")
		return db.UpdateCoursePhaseParticipationParams{}, err
	}

	studentReadableData, err := c.StudentReadableData.GetDBModel()
	if err != nil {
		log.Error("failed to create CoursePhaseParticipation DB model from DTO")
		return db.UpdateCoursePhaseParticipationParams{}, err
	}

	return db.UpdateCoursePhaseParticipationParams{
		CourseParticipationID: c.CourseParticipationID,
		PassStatus:            GetPassStatusDBModel(c.PassStatus),
		RestrictedData:        restrictedData,
		StudentReadableData:   studentReadableData,
		CoursePhaseID:         c.CoursePhaseID,
	}, nil
}
