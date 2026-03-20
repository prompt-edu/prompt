package course

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/core/course/courseDTO"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	log "github.com/sirupsen/logrus"
)

// ErrDuplicateCourseIdentifier is returned when a course with the same name and semester tag already exists.
var ErrDuplicateCourseIdentifier = errors.New("a course with this name and semester already exists")

type CourseService struct {
	queries db.Queries
	conn    *pgxpool.Pool
	// use dependency injection for keycloak to allow mocking
	createCourseGroupsAndRoles func(ctx context.Context, courseName, iterationName, userID string) error
}

var CourseServiceSingleton *CourseService

func GetOwnCourseIDs(ctx context.Context, matriculationNumber, universityLogin string) ([]uuid.UUID, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	courses, err := CourseServiceSingleton.queries.GetOwnCourses(ctxWithTimeout, db.GetOwnCoursesParams{
		MatriculationNumber: pgtype.Text{String: matriculationNumber, Valid: true},
		UniversityLogin:     pgtype.Text{String: universityLogin, Valid: true},
	})
	return courses, err
}

func GetAllCourses(ctx context.Context, userRoles map[string]bool) ([]courseDTO.CourseWithPhases, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	var courses []db.Course
	var err error
	// Get all active courses the user is allowed to see
	if userRoles[permissionValidation.PromptAdmin] {
		// get all courses
		courses, err = CourseServiceSingleton.queries.GetAllActiveCoursesAdmin(ctxWithTimeout)
		if err != nil {
			return nil, err
		}
	} else {
		// get restricted courses
		userRolesArray := []string{}
		for key, value := range userRoles {
			if value {
				userRolesArray = append(userRolesArray, key)
			}
		}
		coursesRestricted, err := CourseServiceSingleton.queries.GetAllActiveCoursesRestricted(ctxWithTimeout, userRolesArray)
		if err != nil {
			return nil, err
		}

		for _, course := range coursesRestricted {
			courses = append(courses, db.Course(course))
		}
	}

	// Get all course phases for each course
	dtoCourses := make([]courseDTO.CourseWithPhases, 0, len(courses))
	for _, course := range courses {
		// Get all course phases for the course
		coursePhases, err := GetCoursePhasesForCourseID(ctx, course.ID)
		if err != nil {
			return nil, err
		}

		courseWithPhases, err := courseDTO.GetCourseWithPhasesDTOFromDBModel(course)
		if err != nil {
			return nil, err
		}

		courseWithPhases.CoursePhases = coursePhases

		dtoCourses = append(dtoCourses, courseWithPhases)
	}

	return dtoCourses, nil
}

func GetCoursePhasesForCourseID(ctx context.Context, courseID uuid.UUID) ([]coursePhaseDTO.CoursePhaseSequence, error) {
	// Get all course phases in order
	coursePhasesOrder, err := CourseServiceSingleton.queries.GetCoursePhaseSequence(ctx, courseID)
	if err != nil {
		return nil, err
	}

	// get all coursePhases out of order
	coursePhasesNoOrder, err := CourseServiceSingleton.queries.GetNotOrderedCoursePhases(ctx, courseID)
	if err != nil {
		return nil, err
	}

	coursePhaseDTO, err := coursePhaseDTO.GetCoursePhaseSequenceDTO(coursePhasesOrder, coursePhasesNoOrder)
	if err != nil {
		return nil, err

	}
	return coursePhaseDTO, nil
}

func GetCourseByID(ctx context.Context, id uuid.UUID) (courseDTO.CourseWithPhases, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()
	course, err := CourseServiceSingleton.queries.GetCourse(ctxWithTimeout, id)
	if err != nil {
		return courseDTO.CourseWithPhases{}, err
	}

	// Get all course phases in order
	coursePhasesOrder, err := CourseServiceSingleton.queries.GetCoursePhaseSequence(ctx, id)
	if err != nil {
		return courseDTO.CourseWithPhases{}, err
	}

	// get all coursePhases out of order
	coursePhasesNoOrder, err := CourseServiceSingleton.queries.GetNotOrderedCoursePhases(ctx, id)
	if err != nil {
		return courseDTO.CourseWithPhases{}, err
	}

	coursePhaseDTO, err := coursePhaseDTO.GetCoursePhaseSequenceDTO(coursePhasesOrder, coursePhasesNoOrder)
	if err != nil {
		return courseDTO.CourseWithPhases{}, err

	}

	CourseWithPhases, err := courseDTO.GetCourseWithPhasesDTOFromDBModel(course)
	if err != nil {
		return courseDTO.CourseWithPhases{}, err
	}

	CourseWithPhases.CoursePhases = coursePhaseDTO

	return CourseWithPhases, nil
}

func CreateCourse(ctx context.Context, course courseDTO.CreateCourse, requesterID string) (courseDTO.Course, error) {
	// start transaction to roll back if keycloak failed
	tx, err := CourseServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return courseDTO.Course{}, err
	}
	defer sdkUtils.DeferRollback(tx, ctx)
	qtx := CourseServiceSingleton.queries.WithTx(tx)

	createCourseParams, err := course.GetDBModel()
	if err != nil {
		return courseDTO.Course{}, err
	}

	createCourseParams.ID = uuid.New()
	if createCourseParams.Template {
		createCourseParams.StartDate = pgtype.Date{Valid: false}
		createCourseParams.EndDate = pgtype.Date{Valid: false}
	}
	createdCourse, err := qtx.CreateCourse(ctx, createCourseParams)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return courseDTO.Course{}, ErrDuplicateCourseIdentifier
		}
		return courseDTO.Course{}, err
	}

	// create keycloak roles - also add the requester to the course lecturer role
	err = CourseServiceSingleton.createCourseGroupsAndRoles(ctx, createdCourse.Name, createdCourse.SemesterTag.String, requesterID)
	if err != nil {
		log.Error("Failed to create keycloak roles for course: ", err)
		return courseDTO.Course{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return courseDTO.Course{}, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return courseDTO.GetCourseDTOFromDBModel(createdCourse)
}

func CheckCourseNameExists(ctx context.Context, name, semesterTag string) (bool, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()
	semesterTagParam := pgtype.Text{Valid: false}
	if semesterTag != "" {
		semesterTagParam = pgtype.Text{String: semesterTag, Valid: true}
	}
	return CourseServiceSingleton.queries.CheckCourseNameExists(ctxWithTimeout, db.CheckCourseNameExistsParams{
		Name:        name,
		SemesterTag: semesterTagParam,
	})
}

func UpdateCoursePhaseOrder(ctx context.Context, courseID uuid.UUID, graphUpdate courseDTO.UpdateCoursePhaseGraph) error {
	tx, err := CourseServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer sdkUtils.DeferRollback(tx, ctx)
	qtx := CourseServiceSingleton.queries.WithTx(tx)

	// delete all previous connections
	err = qtx.DeleteCourseGraph(ctx, courseID)
	if err != nil {
		return err
	}

	// create new connections
	for _, graphItem := range graphUpdate.PhaseGraph {
		err = qtx.CreateCourseGraphConnection(ctx, db.CreateCourseGraphConnectionParams{
			FromCoursePhaseID: graphItem.FromCoursePhaseID,
			ToCoursePhaseID:   graphItem.ToCoursePhaseID,
		})
		if err != nil {
			log.Error("Error creating graph connection: ", err)
			return err
		}
	}

	// reset initial phase to not conflict with unique constraint
	if err = qtx.UpdateInitialCoursePhase(ctx, db.UpdateInitialCoursePhaseParams{
		CourseID: courseID,
		ID:       uuid.UUID{},
	}); err != nil {
		return err
	}

	err = qtx.UpdateInitialCoursePhase(ctx, db.UpdateInitialCoursePhaseParams{
		CourseID: courseID,
		ID:       graphUpdate.InitialPhase,
	})
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func GetCoursePhaseGraph(ctx context.Context, courseID uuid.UUID) ([]courseDTO.CoursePhaseGraph, error) {
	graph, err := CourseServiceSingleton.queries.GetCoursePhaseGraph(ctx, courseID)
	if err != nil {
		return nil, err
	}

	dtoGraph := make([]courseDTO.CoursePhaseGraph, 0, len(graph))
	for _, g := range graph {
		dtoGraph = append(dtoGraph, courseDTO.CoursePhaseGraph{
			FromCoursePhaseID: g.FromCoursePhaseID,
			ToCoursePhaseID:   g.ToCoursePhaseID,
		})
	}
	return dtoGraph, nil
}

func GetParticipationDataGraph(ctx context.Context, courseID uuid.UUID) ([]courseDTO.MetaDataGraphItem, error) {
	graph, err := CourseServiceSingleton.queries.GetParticipationDataGraph(ctx, courseID)
	if err != nil {
		return nil, err
	}

	dtoGraph := make([]courseDTO.MetaDataGraphItem, 0, len(graph))
	for _, g := range graph {
		dtoGraph = append(dtoGraph, courseDTO.MetaDataGraphItem{
			FromCoursePhaseID:    g.FromCoursePhaseID,
			ToCoursePhaseID:      g.ToCoursePhaseID,
			FromCoursePhaseDtoID: g.FromCoursePhaseDtoID,
			ToCoursePhaseDtoID:   g.ToCoursePhaseDtoID,
		})
	}
	return dtoGraph, nil
}

func GetPhaseDataGraph(ctx context.Context, courseID uuid.UUID) ([]courseDTO.MetaDataGraphItem, error) {
	graph, err := CourseServiceSingleton.queries.GetPhaseDataGraph(ctx, courseID)
	if err != nil {
		return nil, err
	}

	dtoGraph := make([]courseDTO.MetaDataGraphItem, 0, len(graph))
	for _, g := range graph {
		dtoGraph = append(dtoGraph, courseDTO.MetaDataGraphItem{
			FromCoursePhaseID:    g.FromCoursePhaseID,
			ToCoursePhaseID:      g.ToCoursePhaseID,
			FromCoursePhaseDtoID: g.FromCoursePhaseDtoID,
			ToCoursePhaseDtoID:   g.ToCoursePhaseDtoID,
		})
	}
	return dtoGraph, nil
}

func UpdateParticipationDataGraph(ctx context.Context, courseID uuid.UUID, graphUpdate []courseDTO.MetaDataGraphItem) error {
	tx, err := CourseServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer sdkUtils.DeferRollback(tx, ctx)
	qtx := CourseServiceSingleton.queries.WithTx(tx)

	// delete all previous connections
	err = qtx.DeleteParticipationDataGraphConnections(ctx, courseID)
	if err != nil {
		return err
	}

	// create new connections
	for _, graphItem := range graphUpdate {
		err = qtx.CreateParticipationDataConnection(ctx, db.CreateParticipationDataConnectionParams{
			FromCoursePhaseID:    graphItem.FromCoursePhaseID,
			ToCoursePhaseID:      graphItem.ToCoursePhaseID,
			FromCoursePhaseDtoID: graphItem.FromCoursePhaseDtoID,
			ToCoursePhaseDtoID:   graphItem.ToCoursePhaseDtoID,
		})
		if err != nil {
			log.Error("Error creating graph connection: ", err)
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil

}

func UpdatePhaseDataGraph(ctx context.Context, courseID uuid.UUID, graphUpdate []courseDTO.MetaDataGraphItem) error {
	tx, err := CourseServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer sdkUtils.DeferRollback(tx, ctx)
	qtx := CourseServiceSingleton.queries.WithTx(tx)

	// delete all previous connections
	err = qtx.DeletePhaseDataGraphConnections(ctx, courseID)
	if err != nil {
		return err
	}

	// create new connections
	for _, graphItem := range graphUpdate {
		err = qtx.CreatePhaseDataConnection(ctx, db.CreatePhaseDataConnectionParams{
			FromCoursePhaseID:    graphItem.FromCoursePhaseID,
			ToCoursePhaseID:      graphItem.ToCoursePhaseID,
			FromCoursePhaseDtoID: graphItem.FromCoursePhaseDtoID,
			ToCoursePhaseDtoID:   graphItem.ToCoursePhaseDtoID,
		})
		if err != nil {
			log.Error("Error creating graph connection: ", err)
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil

}

func UpdateCourseArchiveStatus(
	ctx context.Context,
	courseID uuid.UUID,
	archived bool,
) (courseDTO.Course, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	var archivedOn pgtype.Timestamptz
	if archived {
		archivedOn = pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		}
	}

	res, err := CourseServiceSingleton.queries.ArchiveCourse(
		ctxWithTimeout,
		db.ArchiveCourseParams{
			ID:         courseID,
			Archived:   archived,
			ArchivedOn: archivedOn,
		},
	)
	if err != nil {
		log.Error(err)
		return courseDTO.Course{}, errors.New("failed to update course archive status")
	}

	course, err := courseDTO.GetCourseDTOFromDBModel(res)
	if err != nil {
		log.Error(err)
		return courseDTO.Course{}, errors.New("failed to map course dto")
	}

	return course, nil
}

func UpdateCourseData(ctx context.Context, courseID uuid.UUID, courseData courseDTO.UpdateCourseData) error {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	updateCourseParams, err := courseData.GetDBModel()
	if err != nil {
		log.Error(err)
		return errors.New("failed to update course data")
	}

	updateCourseParams.ID = courseID

	err = CourseServiceSingleton.queries.UpdateCourse(ctxWithTimeout, updateCourseParams)
	if err != nil {
		log.Error(err)
		return errors.New("failed to update course data")
	}

	return nil
}

func DeleteCourse(ctx context.Context, courseID uuid.UUID) error {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	err := CourseServiceSingleton.queries.DeleteCourse(ctxWithTimeout, courseID)
	if err != nil {
		log.Error(err)
		return errors.New("failed to delete course")
	}

	return nil
}

func UpdateCourseTemplateStatus(ctx context.Context, courseID uuid.UUID, isTemplate bool) error {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	if isTemplate {
		err := CourseServiceSingleton.queries.MarkCourseAsTemplate(ctxWithTimeout, courseID)
		if err != nil {
			log.Error(err)
			return errors.New("failed to mark course as template")
		}
	} else {
		err := CourseServiceSingleton.queries.UnmarkCourseAsTemplate(ctxWithTimeout, courseID)
		if err != nil {
			log.Error(err)
			return errors.New("failed to unmark course as template")
		}
	}

	return nil
}

func GetTemplateCourses(ctx context.Context, userRoles map[string]bool) ([]courseDTO.Course, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	var courses []db.Course
	var err error
	if userRoles[permissionValidation.PromptAdmin] {
		courses, err = CourseServiceSingleton.queries.GetTemplateCoursesAdmin(ctxWithTimeout)
		if err != nil {
			return nil, err
		}
	} else {
		userRolesArray := []string{}
		for key, value := range userRoles {
			if value {
				userRolesArray = append(userRolesArray, key)
			}
		}
		coursesRestricted, err := CourseServiceSingleton.queries.GetTemplateCoursesRestricted(ctxWithTimeout, userRolesArray)
		if err != nil {
			return nil, err
		}

		for _, course := range coursesRestricted {
			courses = append(courses, db.Course(course))
		}
	}

	dtoCourses := make([]courseDTO.Course, 0, len(courses))
	for _, course := range courses {
		dtoCourse, err := courseDTO.GetCourseDTOFromDBModel(course)
		if err != nil {
			return nil, err
		}
		dtoCourses = append(dtoCourses, dtoCourse)
	}

	return dtoCourses, nil
}

func CheckCourseTemplateStatus(ctx context.Context, courseID uuid.UUID) (bool, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	isTemplate, err := CourseServiceSingleton.queries.CheckCourseTemplateStatus(ctxWithTimeout, courseID)
	if err != nil {
		log.Error(err)
		return false, errors.New("failed to check if course is template")
	}

	return isTemplate, nil
}
