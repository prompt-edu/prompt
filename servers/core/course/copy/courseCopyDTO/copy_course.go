package courseCopyDTO

import (
	"github.com/google/uuid"
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
	// OrgID puts the new course into an org; it is never taken from the source course.
	OrgID *uuid.UUID `json:"orgID,omitempty" extensions:"x-nullable"`
}
