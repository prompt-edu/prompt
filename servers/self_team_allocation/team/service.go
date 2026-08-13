package teams

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/team/teamDTO"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/timeframe"
	log "github.com/sirupsen/logrus"
)

type TeamsService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var TeamsServiceSingleton *TeamsService

type AssignmentService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var AssignmentServiceSingleton *AssignmentService

func GetAllTeams(ctx context.Context, coursePhaseID uuid.UUID) ([]promptTypes.Team, error) {
	dbTeams, err := TeamsServiceSingleton.queries.GetTeamsWithStudentNames(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not get the teams from the database: ", err)
		return nil, errors.New("could not get the teams from the database")
	}
	dtos, err := teamDTO.GetTeamWithFullNameDTOsFromDBModels(dbTeams)
	if err != nil {
		log.Error("could not get the teams from the database: ", err)
		return nil, errors.New("could not get the teams from the database")
	}
	return dtos, nil
}

func GetTeamByID(ctx context.Context, coursePhaseID uuid.UUID, teamID uuid.UUID) (promptTypes.Team, error) {
	dbTeam, err := TeamsServiceSingleton.queries.GetTeamWithStudentNamesByTeamID(ctx, db.GetTeamWithStudentNamesByTeamIDParams{
		ID:            teamID,
		CoursePhaseID: coursePhaseID,
	})
	if err != nil {
		log.Error("could not get the teams from the database: ", err)
		return promptTypes.Team{}, errors.New("could not get the teams from the database")
	}
	dto, err := teamDTO.GetTeamWithFullNamesByIdDTOFromDBModel(dbTeam)
	if err != nil {
		log.Error("could not get the teams from the database: ", err)
		return promptTypes.Team{}, errors.New("could not get the teams from the database")
	}
	return dto, nil
}

func CreateNewTeams(ctx context.Context, teamNames []string, coursePhaseID uuid.UUID) error {
	// Validate team names
	for _, name := range teamNames {
		if name == "" {
			return errors.New("team name cannot be empty")
		}
	}

	tx, err := TeamsServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := TeamsServiceSingleton.queries.WithTx(tx)

	for _, teamName := range teamNames {
		err := qtx.CreateTeam(ctx, db.CreateTeamParams{
			ID:            uuid.New(),
			Name:          teamName,
			CoursePhaseID: coursePhaseID,
		})
		if err != nil {
			log.Error("error creating the teams ", err)
			return errors.New("error creating the teams")
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func UpdateTeam(ctx context.Context, coursePhaseID, teamID uuid.UUID, newTeamName string) error {
	err := TeamsServiceSingleton.queries.UpdateTeam(ctx, db.UpdateTeamParams{
		ID:            teamID,
		CoursePhaseID: coursePhaseID,
		Name:          newTeamName,
	})
	if err != nil {
		log.Error("could not update the team: ", err)
		return errors.New("could not update the team")
	}
	return nil
}

func AssignTeam(ctx context.Context, coursePhaseID, teamID uuid.UUID, courseParticipationID uuid.UUID, studentFirstName, studentLastName string) error {
	err := AssignmentServiceSingleton.queries.CreateOrUpdateAssignment(ctx, db.CreateOrUpdateAssignmentParams{
		ID:                    uuid.New(),
		TeamID:                teamID,
		CoursePhaseID:         coursePhaseID,
		CourseParticipationID: courseParticipationID,
		StudentFirstName:      studentFirstName,
		StudentLastName:       studentLastName,
	})

	if err != nil {
		log.Error("could not assign student to team: ", err)
		return errors.New("could not assign student to team")
	}
	return nil
}

func LeaveTeam(ctx context.Context, coursePhaseID, teamID uuid.UUID, courseParticipationID uuid.UUID) error {
	err := AssignmentServiceSingleton.queries.DeleteAssignment(ctx, db.DeleteAssignmentParams{
		TeamID:                teamID,
		CoursePhaseID:         coursePhaseID,
		CourseParticipationID: courseParticipationID,
	})

	if err != nil {
		log.Error("could not leave the team: ", err)
		return errors.New("could not leave the team")
	}
	return nil
}

func DeleteTeam(ctx context.Context, coursePhaseID, teamID uuid.UUID) error {
	err := TeamsServiceSingleton.queries.DeleteTeam(ctx, db.DeleteTeamParams{
		ID:            teamID,
		CoursePhaseID: coursePhaseID,
	})
	if err != nil {
		log.Error("could not delete the team: ", err)
		return errors.New("could not delete the team")
	}
	return nil
}

func ValidateTimeframe(ctx context.Context, coursePhaseID uuid.UUID) (bool, error) {
	timeframeDTO, err := timeframe.GetTimeframe(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not get the timeframe: ", err)
		return false, errors.New("could not get the timeframe")
	}

	if !timeframeDTO.TimeframeSet {
		log.Warn("timeframe not set")
		return true, nil
	}

	startTime := time.Date(
		timeframeDTO.StartTime.Year(),
		timeframeDTO.StartTime.Month(),
		timeframeDTO.StartTime.Day(),
		timeframeDTO.StartTime.Hour(),
		timeframeDTO.StartTime.Minute(),
		timeframeDTO.StartTime.Second(),
		timeframeDTO.StartTime.Nanosecond(),
		time.Local,
	)
	endTime := time.Date(
		timeframeDTO.EndTime.Year(),
		timeframeDTO.EndTime.Month(),
		timeframeDTO.EndTime.Day(),
		timeframeDTO.EndTime.Hour(),
		timeframeDTO.EndTime.Minute(),
		timeframeDTO.EndTime.Second(),
		timeframeDTO.EndTime.Nanosecond(),
		time.Local,
	)

	now := time.Now()
	if now.Before(startTime) || now.After(endTime) {
		return false, errors.New("request is outside the allowed timeframe")
	}
	return true, nil
}

func ImportTutors(ctx context.Context, coursePhaseID uuid.UUID, tutors []teamDTO.Tutor) error {
	tx, err := TeamsServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := TeamsServiceSingleton.queries.WithTx(tx)

	for _, tutor := range tutors {
		// store tutor in database
		err := qtx.CreateTutor(ctx, db.CreateTutorParams{
			CoursePhaseID:         coursePhaseID,
			CourseParticipationID: tutor.CourseParticipationID,
			FirstName:             tutor.FirstName,
			LastName:              tutor.LastName,
			TeamID:                tutor.TeamID,
		})
		if err != nil {
			log.Error("could not create tutor: ", err)
			return errors.New("could not create tutor")
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func CreateManualTutor(ctx context.Context, coursePhaseID uuid.UUID, firstName, lastName string, teamID uuid.UUID) error {
	// Generate a new UUID for the course participation ID
	courseParticipationID := uuid.New()

	err := TeamsServiceSingleton.queries.CreateTutor(ctx, db.CreateTutorParams{
		CoursePhaseID:         coursePhaseID,
		CourseParticipationID: courseParticipationID,
		FirstName:             firstName,
		LastName:              lastName,
		TeamID:                teamID,
	})

	if err != nil {
		log.Error("could not create manual tutor: ", err)
		return errors.New("could not create manual tutor")
	}
	return nil
}

func DeleteTutor(ctx context.Context, coursePhaseID, tutorID uuid.UUID) error {
	err := TeamsServiceSingleton.queries.DeleteTutor(ctx, db.DeleteTutorParams{
		CourseParticipationID: tutorID,
		CoursePhaseID:         coursePhaseID,
	})

	if err != nil {
		log.Error("could not delete tutor: ", err)
		return errors.New("could not delete tutor")
	}
	return nil
}

func GetTutorsByCoursePhase(ctx context.Context, coursePhaseID uuid.UUID) ([]teamDTO.Tutor, error) {
	dbTutors, err := TeamsServiceSingleton.queries.GetTutorsByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not get tutors from database: ", err)
		return nil, errors.New("could not get tutors from database")
	}
	return teamDTO.GetTutorDTOsFromDBModels(dbTutors), nil
}
