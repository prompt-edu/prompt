package tease

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/team_allocation/coreRequests"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/team_allocation/survey"
	"github.com/prompt-edu/prompt/servers/team_allocation/tease/teaseDTO"
	log "github.com/sirupsen/logrus"
)

type TeaseService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var TeaseServiceSingleton *TeaseService

func GetTeamAllocationCoursePhases(
	ctx context.Context,
	authHeader string,
	userPermissions map[string]bool,
) ([]teaseDTO.TeasePhase, error) {
	// 1. Request from core to get all relevant courses
	coreURL := sdkUtils.GetCoreUrl()
	courses, err := coreRequests.GetCourses(coreURL, authHeader)
	if err != nil {
		log.Error("could not get courses from core: ", err)
		return nil, err
	}

	// 2. Filter for the courses with a "Team Allocation" course phase and the user has permission
	type RelevantCoursePhase struct {
		CoursePhaseID uuid.UUID
		CourseID      uuid.UUID
		SemesterName  string
	}

	var relevantCoursePhases []RelevantCoursePhase
	for _, course := range courses {
		for _, coursePhase := range course.CoursePhases {
			if coursePhase.CoursePhaseType == "Team Allocation" &&
				hasCoursePhasePermission(userPermissions, course.SemesterTag, course.Name) {

				relevantCoursePhases = append(relevantCoursePhases, RelevantCoursePhase{
					CoursePhaseID: coursePhase.ID,
					CourseID:      coursePhase.CourseID,
					SemesterName:  fmt.Sprintf("%s-%s-%s", course.SemesterTag, course.Name, coursePhase.Name),
				})
			}
		}
	}

	// 3. For each of these course phases, get the survey deadline
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	teasePhases := make([]teaseDTO.TeasePhase, 0, len(relevantCoursePhases))
	for _, coursePhase := range relevantCoursePhases {
		deadline, err := TeaseServiceSingleton.queries.GetSurveyDeadline(ctxWithTimeout, coursePhase.CoursePhaseID)
		if err != nil {
			// Allow missing deadlines but fail for other errors
			if errors.Is(err, sql.ErrNoRows) {
				log.WithFields(log.Fields{
					"course_phase_id": coursePhase.CoursePhaseID,
				}).Warn("no survey deadline found for this course phase, continuing anyway")
			} else {
				log.Error("failed to get survey deadline for course phase: ", err)
				return nil, err
			}
		}

		teasePhases = append(teasePhases, teaseDTO.TeasePhase{
			CoursePhaseID:              coursePhase.CoursePhaseID,
			SemesterName:               coursePhase.SemesterName,
			KickoffSubmissionPeriodEnd: deadline,
		})
	}

	// 4. Return these course phases with names
	return teasePhases, nil
}

func hasCoursePhasePermission(userPermissions map[string]bool, semesterTag, courseName string) bool {
	// Admins always have permission
	if userPermissions[promptSDK.PromptAdmin] {
		return true
	}

	// Otherwise, check for lecturer-specific permission
	requiredPermission := fmt.Sprintf("%s-%s-%s", semesterTag, courseName, promptSDK.CourseLecturer)
	return userPermissions[requiredPermission]
}

func GetTeaseStudentsForCoursePhase(ctx context.Context, authHeader string, coursePhaseID uuid.UUID) ([]teaseDTO.Student, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	coreURL := sdkUtils.GetCoreUrl()
	coursePhaseParticipations, err := promptSDK.FetchAndMergeParticipationsWithResolutions(coreURL, authHeader, coursePhaseID)
	if err != nil {
		log.Error("could not get students from core: ", err)
		return nil, err
	}

	mappedStudents := make([]teaseDTO.Student, 0, len(coursePhaseParticipations))
	for _, cp := range coursePhaseParticipations {
		projectPreferences, err := getTeaseProjectPreferences(ctxWithTimeout, coursePhaseID, cp.CourseParticipationID)
		if err != nil {
			log.Error("could not get project preferences for course participation: ", err)
			return nil, err
		}

		skillResponses, err := getTeaseSkillLevel(ctxWithTimeout, coursePhaseID, cp.CourseParticipationID)
		if err != nil {
			log.Error("could not get skill responses for course participation: ", err)
			return nil, err
		}

		student, err := teaseDTO.ConvertCourseParticipationToTeaseStudent(cp, projectPreferences, skillResponses)
		if err != nil {
			log.Error("could not convert course participation to tease student: ", err)
			return nil, err
		}
		mappedStudents = append(mappedStudents, student)
	}

	return mappedStudents, nil
}

func getTeaseProjectPreferences(ctx context.Context, coursePhaseID uuid.UUID, courseParticipationID uuid.UUID) ([]teaseDTO.ProjectPreference, error) {
	teamType := "standard"
	if survey.GetAllocationProfile(ctx, coursePhaseID) == "project_week_1000_plus" {
		return []teaseDTO.ProjectPreference{}, nil
	}

	teamPreferences, err := TeaseServiceSingleton.queries.GetStudentPreferencesByTeamType(ctx, db.GetStudentPreferencesByTeamTypeParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
		TeamType:              teamType,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Error("could not get team preferences for course participation: ", err)
		return nil, err
	} else if err != nil && errors.Is(err, sql.ErrNoRows) {
		log.WithFields(log.Fields{
			"course_participation_id": courseParticipationID,
			"course_phase_id":         coursePhaseID,
		}).Warn("no team preferences found for this course participation, continuing anyway")
		return []teaseDTO.ProjectPreference{}, nil
	}

	return teaseDTO.GetProjectPreferenceFromTypedDBModel(teamPreferences), nil
}

func getTeaseSkillLevel(ctx context.Context, coursePhaseID uuid.UUID, courseParticipationID uuid.UUID) ([]teaseDTO.StudentSkillResponse, error) {
	preferenceMode := "teams"
	if survey.GetAllocationProfile(ctx, coursePhaseID) == "project_week_1000_plus" {
		preferenceMode = "fields"
	}

	skills, err := TeaseServiceSingleton.queries.GetStudentSkillResponsesByPreferenceMode(ctx, db.GetStudentSkillResponsesByPreferenceModeParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
		PreferenceMode:        preferenceMode,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Error("could not get skills for course participation: ", err)
		return nil, err
	} else if err != nil && errors.Is(err, sql.ErrNoRows) {
		log.WithFields(log.Fields{
			"course_participation_id": courseParticipationID,
			"course_phase_id":         coursePhaseID,
		}).Warn("no skills found for this course participation, continuing anyway")
		return []teaseDTO.StudentSkillResponse{}, nil
	}

	return teaseDTO.GetTeaseStudentSkillResponseFromPreferenceModeDBModel(skills), nil
}

func GetTeaseSkillsByCoursePhase(ctx context.Context, coursePhaseID uuid.UUID) ([]teaseDTO.Skill, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	skills, err := TeaseServiceSingleton.queries.GetSkillsByCoursePhase(ctxWithTimeout, coursePhaseID)
	if err != nil {
		log.Error("failed to get skills for course phase: ", err)
		return nil, err
	}

	return teaseDTO.GetTeaseSkillFromDBModel(skills), nil
}

func GetTeaseTeamsByCoursePhase(ctx context.Context, coursePhaseID uuid.UUID) ([]teaseDTO.TeaseTeam, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	var (
		teams []db.Team
		err   error
	)
	if survey.GetAllocationProfile(ctxWithTimeout, coursePhaseID) == "project_week_1000_plus" {
		rows, queryErr := TeaseServiceSingleton.conn.Query(ctxWithTimeout, `
			SELECT id, name, course_phase_id, created_at, team_type
			FROM team
			WHERE course_phase_id = $1 AND team_type = 'company_project'
			ORDER BY name
		`, coursePhaseID)
		if queryErr != nil {
			log.Error("failed to get teams for course phase: ", queryErr)
			return nil, queryErr
		}
		defer rows.Close()
		for rows.Next() {
			var team db.Team
			if scanErr := rows.Scan(&team.ID, &team.Name, &team.CoursePhaseID, &team.CreatedAt, &team.TeamType); scanErr != nil {
				log.Error("failed to scan teams for course phase: ", scanErr)
				return nil, scanErr
			}
			teams = append(teams, team)
		}
		if err = rows.Err(); err != nil {
			log.Error("failed to iterate teams for course phase: ", err)
			return nil, err
		}
	} else {
		teams, err = TeaseServiceSingleton.queries.GetStandardTeamsByCoursePhase(ctxWithTimeout, coursePhaseID)
	}
	if err != nil {
		log.Error("failed to get teams for course phase: ", err)
		return nil, err
	}

	return teaseDTO.GetTeaseTeamResponseFromDBModel(teams), nil
}

func GetAllocationsByCoursePhase(ctx context.Context, coursePhaseID uuid.UUID) ([]teaseDTO.Allocation, error) {
	rows, err := TeaseServiceSingleton.queries.GetAggregatedAllocationsByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not get the allocations from the database: ", err)
		return nil, fmt.Errorf("could not get the allocations from the database: %w", err)
	}

	teaseAllocations := make([]teaseDTO.Allocation, 0, len(rows))
	for _, row := range rows {
		teaseAllocations = append(teaseAllocations, teaseDTO.Allocation{
			ProjectID: row.TeamID,
			Students:  row.StudentIds,
		})
	}

	return teaseAllocations, nil
}

func PostAllocations(ctx context.Context, allocations []teaseDTO.Allocation, coursePhaseID uuid.UUID) error {
	tx, err := TeaseServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := TeaseServiceSingleton.queries.WithTx(tx)

	for _, allocation := range allocations {
		if allocation.ProjectID == uuid.Nil {
			return fmt.Errorf("invalid project ID in allocation")
		}
		if len(allocation.Students) == 0 {
			log.Warn("allocation with no students: ", allocation.ProjectID)
			continue
		}
		for _, studentID := range allocation.Students {
			if studentID == uuid.Nil {
				return fmt.Errorf("invalid student ID in allocation")
			}
			err = qtx.CreateOrUpdateAllocation(ctx, db.CreateOrUpdateAllocationParams{
				ID:                    uuid.New(),
				CourseParticipationID: studentID,
				TeamID:                allocation.ProjectID,
				CoursePhaseID:         coursePhaseID,
			})
			if err != nil {
				log.Error("could not create or update allocation: ", err)
				return fmt.Errorf("could not create or update allocation: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("transaction commit failed: ", err)
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	return nil
}
