package coursePhaseParticipationDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/meta"
	"github.com/prompt-edu/prompt/servers/core/student/studentDTO"
	log "github.com/sirupsen/logrus"
)

// this version does not contain any restricted data
// and student should also not see the pass status
type CoursePhaseParticipationStudent struct {
	CoursePhaseID         uuid.UUID          `json:"coursePhaseID"`
	CourseParticipationID uuid.UUID          `json:"courseParticipationID"`
	StudentReadableData   meta.MetaData      `json:"studentReadableData"`
	Student               studentDTO.Student `json:"student"`
}

func GetCoursePhaseParticipationStudent(model db.GetCoursePhaseParticipationByUniversityLoginAndCoursePhaseRow) (CoursePhaseParticipationStudent, error) {
	studentReadableData, err := meta.GetMetaDataDTOFromDBModel(model.StudentReadableData)
	if err != nil {
		log.Error("failed to create CoursePhaseParticipation DTO from DB model")
		return CoursePhaseParticipationStudent{}, err
	}

	return CoursePhaseParticipationStudent{
		CoursePhaseID:         model.CoursePhaseID,
		CourseParticipationID: model.CourseParticipationID,
		StudentReadableData:   studentReadableData,
		Student: studentDTO.GetStudentDTOFromDBModel(db.Student{
			ID:                   model.StudentID,
			FirstName:            model.FirstName,
			LastName:             model.LastName,
			Email:                model.Email,
			MatriculationNumber:  model.MatriculationNumber,
			UniversityLogin:      model.UniversityLogin,
			HasUniversityAccount: model.HasUniversityAccount,
			Gender:               model.Gender,
			Nationality:          model.Nationality,
			StudyDegree:          model.StudyDegree,
			StudyProgram:         model.StudyProgram,
			CurrentSemester:      model.CurrentSemester,
		}),
	}, nil
}
