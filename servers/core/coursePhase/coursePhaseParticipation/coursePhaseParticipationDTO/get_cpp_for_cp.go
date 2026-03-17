package coursePhaseParticipationDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/meta"
	"github.com/prompt-edu/prompt/servers/core/student/studentDTO"
	log "github.com/sirupsen/logrus"
)

type GetAllCPPsForCoursePhase struct {
	CoursePhaseID         uuid.UUID          `json:"coursePhaseID"`
	PassStatus            string             `json:"passStatus"`
	CourseParticipationID uuid.UUID          `json:"courseParticipationID"`
	RestrictedData        meta.MetaData      `json:"restrictedData"`
	StudentReadableData   meta.MetaData      `json:"studentReadableData"`
	PrevData              meta.MetaData      `json:"prevData"`
	Student               studentDTO.Student `json:"student"`
}

func GetAllCPPsForCoursePhaseDTOFromDBModel(model db.GetAllCoursePhaseParticipationsForCoursePhaseIncludingPreviousRow) (GetAllCPPsForCoursePhase, error) {
	restrictedData, err := meta.GetMetaDataDTOFromDBModel(model.RestrictedData)
	if err != nil {
		log.Error("failed to create CoursePhaseParticipation DTO from DB model")
		return GetAllCPPsForCoursePhase{}, err
	}

	studentReadableData, err := meta.GetMetaDataDTOFromDBModel(model.StudentReadableData)
	if err != nil {
		log.Error("failed to create CoursePhaseParticipation DTO from DB model")
		return GetAllCPPsForCoursePhase{}, err
	}

	prevData, err := meta.GetMetaDataDTOFromDBModel(model.PrevData)
	if err != nil {
		log.Error("failed to create CoursePhaseParticipation DTO from DB model")
		return GetAllCPPsForCoursePhase{}, err
	}

	return GetAllCPPsForCoursePhase{
		CoursePhaseID:         model.CoursePhaseID,
		CourseParticipationID: model.CourseParticipationID,
		PassStatus:            GetPassStatusString(model.PassStatus),
		RestrictedData:        restrictedData,
		StudentReadableData:   studentReadableData,
		PrevData:              prevData,
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
