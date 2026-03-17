package coursePhaseParticipationDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/meta"
	log "github.com/sirupsen/logrus"
)

type CreateCoursePhaseParticipation struct {
	CourseParticipationID uuid.UUID      `json:"courseParticipationID"`
	CoursePhaseID         uuid.UUID      `json:"coursePhaseID"`
	PassStatus            *db.PassStatus `json:"passStatus"`
	RestrictedData        meta.MetaData  `json:"restrictedData"`
	StudentReadableData   meta.MetaData  `json:"studentReadableData"`
}

func (c CreateCoursePhaseParticipation) GetDBModel() (db.CreateOrUpdateCoursePhaseParticipationParams, error) {
	restrictedDataBytes, err := c.RestrictedData.GetDBModel()
	if err != nil {
		log.Error("failed to create CoursePhaseParticipation DB model from DTO")
		return db.CreateOrUpdateCoursePhaseParticipationParams{}, err
	}

	studentReadableDataBytes, err := c.StudentReadableData.GetDBModel()
	if err != nil {
		log.Error("failed to create CoursePhaseParticipation DB model from DTO")
		return db.CreateOrUpdateCoursePhaseParticipationParams{}, err
	}

	return db.CreateOrUpdateCoursePhaseParticipationParams{
		CourseParticipationID: c.CourseParticipationID,
		CoursePhaseID:         c.CoursePhaseID,
		PassStatus:            GetPassStatusDBModel(c.PassStatus),
		RestrictedData:        restrictedDataBytes,
		StudentReadableData:   studentReadableDataBytes,
	}, nil

}
