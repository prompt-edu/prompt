package infrastructureSetup

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/intro_course/db/sqlc"
	"github.com/prompt-edu/prompt/servers/intro_course/infrastructureSetup/infrastructureDTO"
	log "github.com/sirupsen/logrus"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type InfrastructureService struct {
	queries           db.Queries
	conn              *pgxpool.Pool
	gitlabAccessToken string
}

var InfrastructureServiceSingleton *InfrastructureService

const TOP_LEVEL_GROUP_NAME = "ASE"
const I_PRAKTIKUM_GROUP_NAME = "iPraktikum"

func CreateCourseInfrastructure(coursePhaseID uuid.UUID, semesterTag string) error {
	// 1.) Get Top Level Group
	ipraktikumGroup, err := getiPraktikumGroup()
	if err != nil {
		return err
	}

	courseGroup, err := createCourseIterationGroup(semesterTag, ipraktikumGroup.ID)
	if err != nil {
		log.Error("Failed to create course iteration group: ", err)
		return err
	}

	// 2.) Create the developer group
	_, err = createDeveloperTopLevelGroup(courseGroup.ID)
	if err != nil {
		log.Error("Failed to create developer group: ", err)
	}

	// 3.) Create the tutor groups
	_, err = createTeachingGroup(courseGroup.ID, "tutors")
	if err != nil {
		log.Error("Failed to create tutor group: ", err)
	}

	// 4.) Create the coach group
	_, err = createTeachingGroup(courseGroup.ID, "coaches")
	if err != nil {
		log.Error("Failed to create coach group: ", err)
	}

	// 5.) Create the introCourse group
	_, err = createTeachingGroup(courseGroup.ID, "Introcourse")
	if err != nil {
		log.Error("Failed to create introCourse group: ", err)
	}

	return nil
}

func CreateStudentInfrastructure(ctx context.Context, coursePhaseID, courseParticipationID uuid.UUID, semesterTag, repoName, studentName, submissionDeadline string) error {
	// 1.) get the student developer profile
	devProfile, err := InfrastructureServiceSingleton.queries.GetDeveloperProfileByCourseParticipationID(ctx, db.GetDeveloperProfileByCourseParticipationIDParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("Failed to get developer profile: ", err)
		return errors.New("failed to get developer profile")
	} else if devProfile.GitlabUsername == "" {
		log.Error("cannot create student repo due to missing student gitlab username")
		return errors.New("cannot create student repo due to missing student gitlab username")
	}

	// 2.) Get the assigned tutor
	tutor, err := InfrastructureServiceSingleton.queries.GetAssignedTutor(ctx, db.GetAssignedTutorParams{
		AssignedStudent: pgtype.UUID{Bytes: courseParticipationID, Valid: true},
		CoursePhaseID:   coursePhaseID,
	})
	if err != nil {
		log.Error("Failed to get assigned tutor: ", err)
		return errors.New("failed to get assigned tutor")
	} else if !tutor.GitlabUsername.Valid || tutor.GitlabUsername.String == "" {
		log.Error("cannot create student repo due to missing tutor gitlab username")
		return errors.New("cannot create student repo due to missing tutor gitlab username")
	}

	log.Info("Creating student repo for student: ", devProfile.GitlabUsername, " with tutor: ", tutor.AssignedTutor)

	// 3.) Get Gitlab IDs
	studentGitlabUser, err := getUserID(devProfile.GitlabUsername)
	if err != nil {
		log.Error("Failed to get student gitlab id: ", err)
		return errors.New("failed to get student gitlab id")
	}

	tutorGitlabUser, err := getUserID(tutor.GitlabUsername.String)
	if err != nil {
		log.Error("Failed to get tutor gitlab id: ", err)
		return errors.New("failed to get tutor gitlab id")
	}

	// 4.) Get required GitLab groups
	ipraktikumGroup, err := getiPraktikumGroup()
	if err != nil {
		return err
	}

	semesterGroup, err := getSubGroup(semesterTag, ipraktikumGroup.ID)
	if err != nil {
		log.Error("Failed to get course group: ", err)
		return err
	}

	introCourseGroup, err := getSubGroup("Introcourse", semesterGroup.ID)
	if err != nil {
		log.Error("Failed to get intro course group: ", err)
		return err
	}

	developerGroup, err := getSubGroup("developer", semesterGroup.ID)
	if err != nil {
		log.Error("Failed to get intro course group: ", err)
		return err
	}

	// 5.) Create the student group
	err = CreateStudentProject(repoName, studentGitlabUser.ID, tutorGitlabUser.ID, introCourseGroup.ID, developerGroup.ID, studentName, submissionDeadline)
	if err != nil {
		log.Error("Failed to create student project: ", err)
		// store error in the db
		dbError := InfrastructureServiceSingleton.queries.AddGitlabError(ctx, db.AddGitlabErrorParams{
			CourseParticipationID: courseParticipationID,
			CoursePhaseID:         coursePhaseID,
			ErrorMessage:          pgtype.Text{String: err.Error(), Valid: true},
		})
		if dbError != nil {
			log.Error("Failed to store gitlab error in db: ", dbError)
		}
		return err
	}

	err = InfrastructureServiceSingleton.queries.AddGitlabStatus(ctx, db.AddGitlabStatusParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})

	if err != nil {
		log.Error("Failed to update gitlab status in db: ", err)
		return errors.New("failed to update gitlab status in db")
	}

	return nil
}

func getiPraktikumGroup() (*gitlab.Group, error) {
	ipraktikumGroup, err := getSubGroup(I_PRAKTIKUM_GROUP_NAME, ASE_GROUP_ID)
	if err != nil {
		log.Error("Failed to get group: ", err)
		return nil, err
	}

	return ipraktikumGroup, nil

}

func GetAllStudentGitlabStatus(c context.Context, coursePhaseID uuid.UUID) ([]infrastructureDTO.GitlabStatus, error) {
	// 1.) Get all gitlab status
	gitlabStatuses, err := InfrastructureServiceSingleton.queries.GetAllGitlabStatus(c, coursePhaseID)
	if err != nil {
		log.Error("Failed to get gitlab statuses: ", err)
		return nil, err
	}

	return infrastructureDTO.GetGitlabStatusDTOsFromModels(gitlabStatuses), nil

}

func ManuallyOverwriteStudentGitlabStatus(c context.Context, coursePhaseID, courseParticipationID uuid.UUID) error {
	err := InfrastructureServiceSingleton.queries.AddGitlabStatus(c, db.AddGitlabStatusParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("Failed to update gitlab status in db: ", err)
		return err
	}
	return nil
}
