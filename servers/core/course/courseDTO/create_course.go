package courseDTO

import (
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/meta"
	log "github.com/sirupsen/logrus"
)

type CreateCourse struct {
	Name                string        `json:"name"`
	StartDate           pgtype.Date   `json:"startDate" swaggertype:"string"`
	EndDate             pgtype.Date   `json:"endDate" swaggertype:"string"`
	SemesterTag         pgtype.Text   `json:"semesterTag" swaggertype:"string"`
	RestrictedData      meta.MetaData `json:"restrictedData"`
	StudentReadableData meta.MetaData `json:"studentReadableData"`
	ShortDescription    pgtype.Text   `json:"shortDescription" swaggertype:"string"`
	LongDescription     pgtype.Text   `json:"longDescription" swaggertype:"string"`
	CourseType          db.CourseType `json:"courseType"`
	Ects                pgtype.Int4   `json:"ects" swaggertype:"integer"`
	Template            bool          `json:"template"`
}

func (c CreateCourse) GetDBModel() (db.CreateCourseParams, error) {
	restrictedData, err := c.RestrictedData.GetDBModel()
	if err != nil {
		log.Error("failed to create Course DTO from DB model")
		return db.CreateCourseParams{}, err
	}

	studentReadableData, err := c.StudentReadableData.GetDBModel()
	if err != nil {
		log.Error("failed to create Course DTO from DB model")
		return db.CreateCourseParams{}, err
	}

	return db.CreateCourseParams{
		Name:                c.Name,
		StartDate:           c.StartDate,
		EndDate:             c.EndDate,
		SemesterTag:         c.SemesterTag,
		RestrictedData:      restrictedData,
		StudentReadableData: studentReadableData,
		ShortDescription:    c.ShortDescription,
		LongDescription:     c.LongDescription,
		CourseType:          c.CourseType,
		Ects:                c.Ects,
		Template:            c.Template,
	}, nil
}
