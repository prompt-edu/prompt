package courseParticipation

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt/servers/core/course/courseParticipation/courseParticipationDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/utils"
)

type CourseParticipationService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var CourseParticipationServiceSingleton *CourseParticipationService

func GetAllCourseParticipationsForCourse(ctx context.Context, id uuid.UUID) ([]courseParticipationDTO.GetCourseParticipation, error) {
	courseParticipations, err := CourseParticipationServiceSingleton.queries.GetAllCourseParticipationsForCourse(ctx, id)
	if err != nil {
		return nil, err
	}

	dtoCourseParticipations := make([]courseParticipationDTO.GetCourseParticipation, 0, len(courseParticipations))
	for _, courseParticipation := range courseParticipations {
		dtoCourseParticipation := courseParticipationDTO.GetCourseParticipationDTOFromDBModel(courseParticipation)
		dtoCourseParticipations = append(dtoCourseParticipations, dtoCourseParticipation)
	}

	return dtoCourseParticipations, nil
}

func GetAllCourseParticipationsForStudent(ctx context.Context, id uuid.UUID) ([]courseParticipationDTO.GetCourseParticipation, error) {
	courseParticipations, err := CourseParticipationServiceSingleton.queries.GetAllCourseParticipationsForStudent(ctx, id)
	if err != nil {
		return nil, err
	}

	dtoCourseParticipations := make([]courseParticipationDTO.GetCourseParticipation, 0, len(courseParticipations))
	for _, courseParticipation := range courseParticipations {
		dtoCourseParticipation := courseParticipationDTO.GetCourseParticipationDTOFromDBModel(courseParticipation)
		dtoCourseParticipations = append(dtoCourseParticipations, dtoCourseParticipation)
	}

	return dtoCourseParticipations, nil
}

func CreateCourseParticipation(ctx context.Context, transactionQueries *db.Queries, c courseParticipationDTO.CreateCourseParticipation) (courseParticipationDTO.GetCourseParticipation, error) {
	queries := utils.GetQueries(transactionQueries, &CourseParticipationServiceSingleton.queries)
	courseParticipation := c.GetDBModel()
	courseParticipation.ID = uuid.New()

	createdParticipation, err := queries.CreateCourseParticipation(ctx, courseParticipation)
	if err != nil {
		return courseParticipationDTO.GetCourseParticipation{}, err
	}

	return courseParticipationDTO.GetCourseParticipationDTOFromDBModel(createdParticipation), nil
}

func CreateIfNotExistingCourseParticipation(ctx context.Context, transactionQueries *db.Queries, studentID uuid.UUID, courseID uuid.UUID) (courseParticipationDTO.GetCourseParticipation, error) {
	queries := utils.GetQueries(transactionQueries, &CourseParticipationServiceSingleton.queries)
	participation, err := queries.GetCourseParticipationByStudentAndCourseID(ctx, db.GetCourseParticipationByStudentAndCourseIDParams{
		StudentID: studentID,
		CourseID:  courseID,
	})
	if err == nil {
		return courseParticipationDTO.GetCourseParticipationDTOFromDBModel(participation), nil
	} else if errors.Is(err, sql.ErrNoRows) {
		// has to be created
		return CreateCourseParticipation(ctx, &queries, courseParticipationDTO.CreateCourseParticipation{
			StudentID: studentID,
			CourseID:  courseID,
		})

	} else {
		return courseParticipationDTO.GetCourseParticipation{}, err
	}
}

func GetOwnCourseParticipation(ctx context.Context, courseId uuid.UUID, matriculationNumber, universityLogin string) (courseParticipationDTO.GetOwnCourseParticipation, error) {
	participation, err := CourseParticipationServiceSingleton.queries.GetCourseParticipationByCourseIDAndMatriculation(ctx, db.GetCourseParticipationByCourseIDAndMatriculationParams{
		CourseID:            courseId,
		MatriculationNumber: pgtype.Text{String: matriculationNumber, Valid: true},
		UniversityLogin:     pgtype.Text{String: universityLogin, Valid: true},
	})
	if errors.Is(err, sql.ErrNoRows) {
		return courseParticipationDTO.GetOwnCourseParticipation{
			IsStudentOfCourse: false,
		}, nil
	} else if err != nil {
		return courseParticipationDTO.GetOwnCourseParticipation{}, err
	}

	return courseParticipationDTO.GetOwnCourseParticipationDTOFromDBModel(participation), nil
}
