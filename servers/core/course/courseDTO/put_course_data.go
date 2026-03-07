package courseDTO

import (
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/meta"
	log "github.com/sirupsen/logrus"
)

type UpdateCourseData struct {
	StartDate           pgtype.Date   `json:"startDate" swaggertype:"string"`
	EndDate             pgtype.Date   `json:"endDate" swaggertype:"string"`
	Ects                pgtype.Int4   `json:"ects" swaggertype:"integer"`
	CourseType          pgtype.Text   `json:"courseType" swaggertype:"string"`
	RestrictedData      meta.MetaData `json:"restrictedData"`
	StudentReadableData meta.MetaData `json:"studentReadableData"`
	ShortDescription    pgtype.Text   `json:"shortDescription" swaggertype:"string"`
	LongDescription     pgtype.Text   `json:"longDescription" swaggertype:"string"`
}

func (c UpdateCourseData) GetDBModel() (db.UpdateCourseParams, error) {
	restrictedData, err := c.RestrictedData.GetDBModel()
	if err != nil {
		log.Error("failed to create Course DTO from DB model")
		return db.UpdateCourseParams{}, err
	}

	studentReadableData, err := c.StudentReadableData.GetDBModel()
	if err != nil {
		log.Error("failed to create Course DTO from DB model")
		return db.UpdateCourseParams{}, err
	}

	var nullCourseType db.NullCourseType
	if c.CourseType.Valid {
		courseType := db.CourseType(c.CourseType.String)

		nullCourseType = db.NullCourseType{
			CourseType: courseType,
			Valid:      true,
		}
	} else {
		nullCourseType = db.NullCourseType{
			Valid: false,
		}
	}

	return db.UpdateCourseParams{
		RestrictedData:      restrictedData,
		StudentReadableData: studentReadableData,
		StartDate:           c.StartDate,
		EndDate:             c.EndDate,
		Ects:                c.Ects,
		CourseType:          nullCourseType,
		ShortDescription:    c.ShortDescription,
		LongDescription:     c.LongDescription,
	}, nil
}
