package teams

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/team_allocation/team/teamDTO"
	log "github.com/sirupsen/logrus"
)

var (
	ErrTutorNotFound       = errors.New("tutor not found")
	ErrDuplicateLogin      = errors.New("university login already assigned to another tutor in this phase")
	ErrInvalidTeamForPhase = errors.New("team does not belong to this course phase")
)

type TeamsService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var TeamsServiceSingleton *TeamsService

func GetAllTeams(ctx context.Context, coursePhaseID uuid.UUID) ([]promptTypes.Team, error) {
	dbTeams, err := TeamsServiceSingleton.queries.GetTeamsWithMembers(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not get the teams from the database: ", err)
		return nil, errors.New("could not get the teams from the database")
	}
	dtos, err := teamDTO.GetTeamsWithMembersDTOFromDBModel(dbTeams)
	if err != nil {
		log.Error("could not get the teams from the database: ", err)
		return nil, errors.New("could not get the teams from the database")
	}
	return dtos, nil
}

func GetTeamByID(ctx context.Context, coursePhaseID uuid.UUID, teamID uuid.UUID) (promptTypes.Team, error) {
	dbTeam, err := TeamsServiceSingleton.queries.GetTeamByCoursePhaseAndTeamID(ctx, db.GetTeamByCoursePhaseAndTeamIDParams{
		ID:            teamID,
		CoursePhaseID: coursePhaseID,
	})
	if err != nil {
		log.Error("could not get the teams from the database: ", err)
		return promptTypes.Team{}, fmt.Errorf("could not get the teams from the database: %w", err)
	}
	return teamDTO.GetTeamDTOFromDBModel(dbTeam), nil
}

func CreateNewTeams(ctx context.Context, teamNames []string, coursePhaseID uuid.UUID) error {
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

func AddStudentNamesToAllocations(ctx context.Context, req teamDTO.StudentNameUpdateRequest) error {
	tx, err := TeamsServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := TeamsServiceSingleton.queries.WithTx(tx)

	for participationID, name := range req.StudentNamesPerID {
		err := qtx.UpdateStudentNameForAllocation(ctx, db.UpdateStudentNameForAllocationParams{
			StudentFirstName:      name.FirstName,
			StudentLastName:       name.LastName,
			CourseParticipationID: participationID,
			CoursePhaseID:         req.CoursePhaseID,
		})
		if err != nil {
			return fmt.Errorf("failed to update name for participation ID %s: %w", participationID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func ImportTutors(ctx context.Context, coursePhaseID uuid.UUID, tutors []teamDTO.Tutor) error {
	tx, err := TeamsServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer sdkUtils.DeferRollback(tx, ctx)
	qtx := TeamsServiceSingleton.queries.WithTx(tx)

	for _, tutor := range tutors {
		normalizedLogin := teamDTO.NormalizeUniversityLogin(tutor.UniversityLogin)
		err := qtx.UpsertTutor(ctx, db.UpsertTutorParams{
			CoursePhaseID:         coursePhaseID,
			CourseParticipationID: tutor.CourseParticipationID,
			FirstName:             tutor.FirstName,
			LastName:              tutor.LastName,
			TeamID:                tutor.TeamID,
			UniversityLogin:       teamDTO.UniversityLoginParam(normalizedLogin),
		})
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return fmt.Errorf("%q: %w", normalizedLogin, ErrDuplicateLogin)
			}
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func UpdateTutorTeam(ctx context.Context, coursePhaseID uuid.UUID, universityLogin string, teamID uuid.UUID) error {
	rows, err := TeamsServiceSingleton.queries.UpdateTutorTeam(ctx, db.UpdateTutorTeamParams{
		CoursePhaseID:   coursePhaseID,
		UniversityLogin: teamDTO.UniversityLoginParam(universityLogin),
		TeamID:          teamID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ErrInvalidTeamForPhase
		}
		return err
	}
	if rows == 0 {
		return ErrTutorNotFound
	}
	return nil
}
