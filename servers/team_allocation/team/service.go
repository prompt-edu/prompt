package teams

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt-sdk/tutorscope"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/team_allocation/team/teamDTO"
	log "github.com/sirupsen/logrus"
)

var (
	ErrMissingLogin        = errors.New("every tutor needs a university login")
	ErrMissingTutorID      = errors.New("every tutor needs a courseParticipationID")
	ErrDuplicateLogin      = errors.New("university login is listed more than once")
	ErrNoTeams             = errors.New("tutor needs at least one team")
	ErrInvalidTeamForPhase = errors.New("team does not belong to this course phase")
)

type TeamsService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

func NewTeamsService(queries db.Queries, conn *pgxpool.Pool) *TeamsService {
	return &TeamsService{
		queries: queries,
		conn:    conn,
	}
}

func (s *TeamsService) GetAllTeams(ctx context.Context, coursePhaseID uuid.UUID) ([]promptTypes.Team, error) {
	dbTeams, err := s.queries.GetTeamsWithMembers(ctx, coursePhaseID)
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

func (s *TeamsService) GetTeamByID(ctx context.Context, coursePhaseID uuid.UUID, teamID uuid.UUID) (promptTypes.Team, error) {
	dbTeam, err := s.queries.GetTeamByCoursePhaseAndTeamID(ctx, db.GetTeamByCoursePhaseAndTeamIDParams{
		ID:            teamID,
		CoursePhaseID: coursePhaseID,
	})
	if err != nil {
		log.Error("could not get the teams from the database: ", err)
		return promptTypes.Team{}, fmt.Errorf("could not get the teams from the database: %w", err)
	}
	return teamDTO.GetTeamDTOFromDBModel(dbTeam), nil
}

func (s *TeamsService) CreateNewTeams(ctx context.Context, teamNames []string, coursePhaseID uuid.UUID) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

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

func (s *TeamsService) UpdateTeam(ctx context.Context, coursePhaseID, teamID uuid.UUID, newTeamName string) error {
	err := s.queries.UpdateTeam(ctx, db.UpdateTeamParams{
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

func (s *TeamsService) DeleteTeam(ctx context.Context, coursePhaseID, teamID uuid.UUID) error {
	err := s.queries.DeleteTeam(ctx, db.DeleteTeamParams{
		ID:            teamID,
		CoursePhaseID: coursePhaseID,
	})
	if err != nil {
		log.Error("could not delete the team: ", err)
		return errors.New("could not delete the team")
	}
	return nil
}

func (s *TeamsService) AddStudentNamesToAllocations(ctx context.Context, coursePhaseID uuid.UUID, req teamDTO.StudentNameUpdateRequest) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	for participationID, name := range req.StudentNamesPerID {
		err := qtx.UpdateStudentNameForAllocation(ctx, db.UpdateStudentNameForAllocationParams{
			StudentFirstName:      name.FirstName,
			StudentLastName:       name.LastName,
			CourseParticipationID: participationID,
			CoursePhaseID:         coursePhaseID,
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

// ImportTutors sets the teams of every listed tutor. A tutor is identified by their university
// login: their previous assignments in the phase are replaced, and a tutor who already exists keeps
// the id they carry, so evaluations other phases stored against it stay attached.
func (s *TeamsService) ImportTutors(ctx context.Context, coursePhaseID uuid.UUID, tutors []teamDTO.Tutor) error {
	if err := validateTutorImport(tutors); err != nil {
		return err
	}

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer sdkUtils.DeferRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	for _, tutor := range tutors {
		login := teamDTO.UniversityLoginParam(tutorscope.NormalizeLogin(tutor.UniversityLogin))

		tutorID, err := qtx.GetTutorIDByLogin(ctx, db.GetTutorIDByLoginParams{
			CoursePhaseID:   coursePhaseID,
			UniversityLogin: login,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			tutorID = tutor.CourseParticipationID
		} else if err != nil {
			return err
		}

		if err := qtx.DeleteTutorAssignments(ctx, db.DeleteTutorAssignmentsParams{
			CoursePhaseID:         coursePhaseID,
			UniversityLogin:       login,
			CourseParticipationID: tutorID,
		}); err != nil {
			return err
		}

		for _, teamID := range uniqueTeamIDs(tutor.TeamIDs) {
			err := qtx.InsertTutor(ctx, db.InsertTutorParams{
				CoursePhaseID:         coursePhaseID,
				CourseParticipationID: tutorID,
				FirstName:             tutor.FirstName,
				LastName:              tutor.LastName,
				TeamID:                teamID,
				UniversityLogin:       login,
			})
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return ErrInvalidTeamForPhase
			}
			if err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func validateTutorImport(tutors []teamDTO.Tutor) error {
	seenLogins := make(map[string]bool, len(tutors))
	for _, tutor := range tutors {
		login := tutorscope.NormalizeLogin(tutor.UniversityLogin)
		if login == "" {
			return ErrMissingLogin
		}
		if seenLogins[login] {
			return fmt.Errorf("%q: %w", login, ErrDuplicateLogin)
		}
		seenLogins[login] = true

		if tutor.CourseParticipationID == uuid.Nil {
			return ErrMissingTutorID
		}
		if len(tutor.TeamIDs) == 0 {
			return fmt.Errorf("%q: %w", login, ErrNoTeams)
		}
	}
	return nil
}

func uniqueTeamIDs(teamIDs []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]bool, len(teamIDs))
	unique := make([]uuid.UUID, 0, len(teamIDs))
	for _, teamID := range teamIDs {
		if !seen[teamID] {
			seen[teamID] = true
			unique = append(unique, teamID)
		}
	}
	return unique
}
