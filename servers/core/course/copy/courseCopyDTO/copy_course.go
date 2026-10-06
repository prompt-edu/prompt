package courseCopyDTO

import (
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type CopyCourseRequest struct {
	Name             string        `json:"name"`
	SemesterTag      pgtype.Text   `json:"semesterTag"`
	StartDate        pgtype.Date   `json:"startDate"`
	EndDate          pgtype.Date   `json:"endDate"`
	ShortDescription pgtype.Text   `json:"shortDescription"`
	LongDescription  pgtype.Text   `json:"longDescription"`
	CourseType       db.CourseType `json:"courseType"`
	Ects             pgtype.Int4   `json:"ects"`
	Template         bool          `json:"template"`
}
