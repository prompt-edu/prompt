package studentDTO

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

// for StudentWithCourses query
type StudentCourseParticipationDTO struct {
	CourseID            uuid.UUID      `json:"courseId"`
	CourseName          string         `json:"courseName"`
	StudentReadableData map[string]any `json:"studentReadableData"`
}

type StudentNoteTagDTO struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color string    `json:"color"`
}

type StudentWithCourseParticipationsDTO struct {
	ID                   uuid.UUID      `json:"id"`
	FirstName            string         `json:"firstName"`
	LastName             string         `json:"lastName"`
	Email                string         `json:"email"`
	HasUniversityAccount bool           `json:"hasUniversityAccount"`
	CurrentSemester      pgtype.Int4    `json:"currentSemester" swaggertype:"integer"`
	Gender               db.Gender      `json:"gender"`
	Nationality          string         `json:"nationality"`
	StudyDegree          db.StudyDegree `json:"studyDegree"`
	StudyProgram         string         `json:"studyProgram"`
	LastModified         time.Time      `json:"lastModified"`

	Courses  []StudentCourseParticipationDTO `json:"courses"`
	NoteTags []StudentNoteTagDTO             `json:"noteTags"`
}

func GetStudentWithCoursesFromDB(row db.GetAllStudentsWithCourseParticipationsRow) (StudentWithCourseParticipationsDTO, error) {
	var courses []StudentCourseParticipationDTO
	if err := json.Unmarshal(row.Courses, &courses); err != nil {
		return StudentWithCourseParticipationsDTO{}, err
	}

	var noteTags []StudentNoteTagDTO
	if err := json.Unmarshal(row.NoteTags, &noteTags); err != nil {
		return StudentWithCourseParticipationsDTO{}, err
	}

	return StudentWithCourseParticipationsDTO{
		ID:                   row.StudentID,
		FirstName:            row.StudentFirstName.String,
		LastName:             row.StudentLastName.String,
		Email:                row.StudentEmail.String,
		HasUniversityAccount: row.StudentHasUniversityAccount.Bool,
		CurrentSemester:      row.StudentCurrentSemester,
		Gender:               row.StudentGender,
		Nationality:          row.StudentNationality.String,
		StudyDegree:          row.StudentStudyDegree,
		StudyProgram:         row.StudentStudyProgram.String,
		LastModified:         row.StudentLastModified.Time,
		Courses:              courses,
		NoteTags:             noteTags,
	}, nil
}
