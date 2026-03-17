package applicationDTO

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type OpenApplication struct {
	CourseName               string      `json:"courseName"`
	CoursePhaseID            uuid.UUID   `json:"id"`
	CourseType               string      `json:"courseType"`
	ECTS                     int         `json:"ects"`
	StartDate                pgtype.Date `json:"startDate" swaggertype:"string"`
	EndDate                  pgtype.Date `json:"endDate" swaggertype:"string"`
	ApplicationDeadline      string      `json:"applicationDeadline"`
	ExternalStudentsAllowed  bool        `json:"externalStudentsAllowed"`
	UniversityLoginAvailable bool        `json:"universityLoginAvailable"`
	ShortDescription         *string     `json:"shortDescription,omitempty"`
	LongDescription          *string     `json:"longDescription,omitempty"`
}

func GetOpenApplicationPhaseDTO(dbModel db.GetAllOpenApplicationPhasesRow) OpenApplication {
	var shortDesc *string
	if dbModel.ShortDescription.Valid {
		shortDesc = &dbModel.ShortDescription.String
	}

	var longDesc *string
	if dbModel.LongDescription.Valid {
		longDesc = &dbModel.LongDescription.String
	}

	return OpenApplication{
		CourseName:               dbModel.CourseName,
		CoursePhaseID:            dbModel.CoursePhaseID,
		CourseType:               string(dbModel.CourseType),
		ECTS:                     int(dbModel.Ects.Int32),
		StartDate:                dbModel.StartDate,
		EndDate:                  dbModel.EndDate,
		ApplicationDeadline:      dbModel.ApplicationEndDate,
		ExternalStudentsAllowed:  dbModel.ExternalStudentsAllowed,
		UniversityLoginAvailable: dbModel.UniversityLoginAvailable,
		ShortDescription:         shortDesc,
		LongDescription:          longDesc,
	}
}
