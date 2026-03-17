package studentDTO

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CoursePhaseTypeDTO struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type CoursePhaseEnrollmentDTO struct {
	CoursePhaseID   uuid.UUID          `json:"coursePhaseId"`
	Name            string             `json:"name"`
	IsInitialPhase  bool               `json:"isInitialPhase"`
	CoursePhaseType CoursePhaseTypeDTO `json:"coursePhaseType"`
	PassStatus      *string            `json:"passStatus"`
	LastModified    *string            `json:"lastModified"`
}

type CourseEnrollmentDTO struct {
	CourseID              uuid.UUID                  `json:"courseId"`
	CourseParticipationID *uuid.UUID                 `json:"courseParticipationId"`
	Name                  string                     `json:"name"`
	SemesterTag           string                     `json:"semesterTag"`
	CourseType            string                     `json:"courseType"`
	Ects                  int32                      `json:"ects"`
	StartDate             *time.Time                 `json:"startDate"`
	EndDate               *time.Time                 `json:"endDate"`
	LongDescription       *string                    `json:"longDescription"`
	StudentReadableData   json.RawMessage            `json:"studentReadableData"`
	CoursePhases          []CoursePhaseEnrollmentDTO `json:"coursePhases"`
}

type StudentEnrollmentsDTO struct {
	Courses []CourseEnrollmentDTO `json:"courses"`
}

func GetStudentEnrollmentsDTOFromDB(row []byte) (StudentEnrollmentsDTO, error) {
	if len(row) == 0 || string(row) == "null" {
		return StudentEnrollmentsDTO{Courses: []CourseEnrollmentDTO{}}, nil
	}

	var courses []CourseEnrollmentDTO
	if err := json.Unmarshal(row, &courses); err != nil {
		return StudentEnrollmentsDTO{}, err
	}

	return StudentEnrollmentsDTO{
		Courses: courses,
	}, nil
}
