package teams

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt-sdk/tutorscope"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/team/teamDTO"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/timeframe/timeframeDTO"
	log "github.com/sirupsen/logrus"
)

const (
	// MaxTeamNameLength bounds team names in characters. The column allows 255, and the
	// client input enforces the same limit as here.
	MaxTeamNameLength = 100
	// MaxTeamsCreatedPerStudent caps how many teams a single student can create in a phase.
	// Students cannot delete teams, so the cap keeps one student from flooding the phase.
	MaxTeamsCreatedPerStudent = 2
	// MaxTeamsPerPhase stops students from creating further teams once a phase holds this
	// many. Lecturers are not limited by it.
	MaxTeamsPerPhase = 200

	uniqueViolationCode     = "23505"
	foreignKeyViolationCode = "23503"
)

var (
	ErrInvalidTeamName   = errors.New("invalid team name")
	ErrDuplicateTeamName = errors.New("a team with this name already exists")
	ErrTeamLimitReached  = errors.New("team creation limit reached")
	ErrOutsideTimeframe  = errors.New("request is outside the allowed timeframe")

	ErrMissingLogin        = errors.New("every tutor needs a university login")
	ErrMissingTutorID      = errors.New("every tutor needs a courseParticipationID")
	ErrDuplicateLogin      = errors.New("university login is listed more than once")
	ErrDuplicateTutorID    = errors.New("courseParticipationID is listed for more than one login")
	ErrNoTeams             = errors.New("tutor needs at least one team")
	ErrInvalidTeamForPhase = errors.New("team does not belong to this course phase")
	ErrTutorNotFound       = errors.New("tutor not found")
)

type timeframeProvider interface {
	GetTimeframe(ctx context.Context, coursePhaseID uuid.UUID) (timeframeDTO.Timeframe, error)
}

type TeamsService struct {
	queries   db.Queries
	conn      *pgxpool.Pool
	timeframe timeframeProvider
}

func NewTeamsService(queries db.Queries, conn *pgxpool.Pool, timeframe timeframeProvider) *TeamsService {
	return &TeamsService{
		queries:   queries,
		conn:      conn,
		timeframe: timeframe,
	}
}

// MaxTeamSize is the number of students a self-allocated team can hold.
const MaxTeamSize = 3

var (
	ErrTeamFull     = errors.New("team is already full")
	ErrTeamNotFound = errors.New("team not found")
)

type AssignmentService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

func NewAssignmentService(queries db.Queries, conn *pgxpool.Pool) *AssignmentService {
	return &AssignmentService{
		queries: queries,
		conn:    conn,
	}
}

func (s *TeamsService) GetAllTeams(ctx context.Context, coursePhaseID uuid.UUID) ([]promptTypes.Team, error) {
	dbTeams, err := s.queries.GetTeamsWithStudentNames(ctx, coursePhaseID)
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

func (s *TeamsService) GetTeamByID(ctx context.Context, coursePhaseID uuid.UUID, teamID uuid.UUID) (promptTypes.Team, error) {
	dbTeam, err := s.queries.GetTeamWithStudentNamesByTeamID(ctx, db.GetTeamWithStudentNamesByTeamIDParams{
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

// CreateNewTeams creates the given teams in a course phase. creatorCourseParticipationID is the
// creating student, or uuid.Nil when a lecturer creates the teams. Student creations are
// recorded and limited per student and per phase.
func (s *TeamsService) CreateNewTeams(ctx context.Context, teamNames []string, coursePhaseID, creatorCourseParticipationID uuid.UUID) error {
	names, err := normalizeTeamNames(teamNames)
	if err != nil {
		return err
	}

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	var creator pgtype.UUID
	if creatorCourseParticipationID != uuid.Nil {
		creator = pgtype.UUID{Bytes: creatorCourseParticipationID, Valid: true}
		if err := checkStudentCreationLimits(ctx, qtx, coursePhaseID, creatorCourseParticipationID, len(names)); err != nil {
			return err
		}
	}

	for _, teamName := range names {
		err := qtx.CreateTeam(ctx, db.CreateTeamParams{
			ID:                             uuid.New(),
			Name:                           teamName,
			CoursePhaseID:                  coursePhaseID,
			CreatedByCourseParticipationID: creator,
		})
		if isUniqueViolation(err) {
			return fmt.Errorf("team %q: %w", teamName, ErrDuplicateTeamName)
		}
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

// checkStudentCreationLimits must run inside the creating transaction: it takes a per-phase
// lock so concurrent requests cannot both pass the counts.
func checkStudentCreationLimits(ctx context.Context, qtx *db.Queries, coursePhaseID, creatorCourseParticipationID uuid.UUID, newTeams int) error {
	if err := qtx.LockTeamCreation(ctx, coursePhaseID); err != nil {
		log.Error("could not lock team creation: ", err)
		return errors.New("error creating the teams")
	}

	createdByStudent, err := qtx.CountTeamsCreatedByParticipation(ctx, db.CountTeamsCreatedByParticipationParams{
		CoursePhaseID:                  coursePhaseID,
		CreatedByCourseParticipationID: pgtype.UUID{Bytes: creatorCourseParticipationID, Valid: true},
	})
	if err != nil {
		log.Error("could not count the teams created by the student: ", err)
		return errors.New("error creating the teams")
	}
	if createdByStudent+int64(newTeams) > MaxTeamsCreatedPerStudent {
		return fmt.Errorf("a student can create at most %d teams: %w", MaxTeamsCreatedPerStudent, ErrTeamLimitReached)
	}

	inPhase, err := qtx.CountTeamsByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not count the teams of the phase: ", err)
		return errors.New("error creating the teams")
	}
	if inPhase+int64(newTeams) > MaxTeamsPerPhase {
		return fmt.Errorf("the phase already has the maximum of %d teams: %w", MaxTeamsPerPhase, ErrTeamLimitReached)
	}
	return nil
}

func normalizeTeamNames(teamNames []string) ([]string, error) {
	if len(teamNames) == 0 {
		return nil, fmt.Errorf("at least one team name is required: %w", ErrInvalidTeamName)
	}
	names := make([]string, 0, len(teamNames))
	seen := make(map[string]struct{}, len(teamNames))
	for _, raw := range teamNames {
		name, err := normalizeTeamName(raw)
		if err != nil {
			return nil, err
		}
		if _, duplicate := seen[name]; duplicate {
			return nil, fmt.Errorf("team %q: %w", name, ErrDuplicateTeamName)
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	return names, nil
}

// normalizeTeamName trims the name and rejects empty, overlong or control-character names.
func normalizeTeamName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", fmt.Errorf("team name cannot be empty: %w", ErrInvalidTeamName)
	}
	if utf8.RuneCountInString(name) > MaxTeamNameLength {
		return "", fmt.Errorf("team name cannot be longer than %d characters: %w", MaxTeamNameLength, ErrInvalidTeamName)
	}
	if strings.ContainsFunc(name, unicode.IsControl) {
		return "", fmt.Errorf("team name cannot contain control characters: %w", ErrInvalidTeamName)
	}
	return name, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode
}

func (s *TeamsService) UpdateTeam(ctx context.Context, coursePhaseID, teamID uuid.UUID, newTeamName string) error {
	name, err := normalizeTeamName(newTeamName)
	if err != nil {
		return err
	}
	err = s.queries.UpdateTeam(ctx, db.UpdateTeamParams{
		ID:            teamID,
		CoursePhaseID: coursePhaseID,
		Name:          name,
	})
	if isUniqueViolation(err) {
		return fmt.Errorf("team %q: %w", name, ErrDuplicateTeamName)
	}
	if err != nil {
		log.Error("could not update the team: ", err)
		return errors.New("could not update the team")
	}
	return nil
}

// AssignTeam puts the student into the team unless it already holds MaxTeamSize members.
// The team row is locked for the count and the insert, so concurrent joins cannot overfill it.
func (s *AssignmentService) AssignTeam(ctx context.Context, coursePhaseID, teamID uuid.UUID, courseParticipationID uuid.UUID, studentFirstName, studentLastName string) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		log.Error("could not begin the team assignment: ", err)
		return errors.New("could not assign student to team")
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	locked, err := qtx.LockTeamForAssignment(ctx, db.LockTeamForAssignmentParams{
		ID:            teamID,
		CoursePhaseID: coursePhaseID,
	})
	if err != nil {
		log.Error("could not lock the team: ", err)
		return errors.New("could not assign student to team")
	}
	if len(locked) == 0 {
		return ErrTeamNotFound
	}

	members, err := qtx.CountTeamMembers(ctx, teamID)
	if err != nil {
		log.Error("could not count the team members: ", err)
		return errors.New("could not assign student to team")
	}
	if members >= MaxTeamSize {
		return ErrTeamFull
	}

	err = qtx.CreateOrUpdateAssignment(ctx, db.CreateOrUpdateAssignmentParams{
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

	if err := tx.Commit(ctx); err != nil {
		log.Error(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (s *AssignmentService) LeaveTeam(ctx context.Context, coursePhaseID, teamID uuid.UUID, courseParticipationID uuid.UUID) error {
	err := s.queries.DeleteAssignment(ctx, db.DeleteAssignmentParams{
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

func (s *TeamsService) ValidateTimeframe(ctx context.Context, coursePhaseID uuid.UUID) (bool, error) {
	timeframe, err := s.timeframe.GetTimeframe(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not get the timeframe: ", err)
		return false, errors.New("could not get the timeframe")
	}

	if !timeframe.TimeframeSet {
		log.Warn("timeframe not set")
		return true, nil
	}

	if time.Now().Before(timeframe.StartTime) || time.Now().After(timeframe.EndTime) {
		return false, ErrOutsideTimeframe
	}
	return true, nil
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
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	if err := qtx.LockTutorImport(ctx, coursePhaseID); err != nil {
		return err
	}

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
			if errors.As(err, &pgErr) && pgErr.Code == foreignKeyViolationCode {
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
	seenIDs := make(map[uuid.UUID]bool, len(tutors))
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
		if seenIDs[tutor.CourseParticipationID] {
			return fmt.Errorf("%s: %w", tutor.CourseParticipationID, ErrDuplicateTutorID)
		}
		seenIDs[tutor.CourseParticipationID] = true
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

func (s *TeamsService) CreateManualTutor(ctx context.Context, coursePhaseID uuid.UUID, firstName, lastName string, teamID uuid.UUID) error {
	// Generate a new UUID for the course participation ID
	courseParticipationID := uuid.New()

	err := s.queries.CreateTutor(ctx, db.CreateTutorParams{
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

func (s *TeamsService) DeleteTutor(ctx context.Context, coursePhaseID, tutorID uuid.UUID) error {
	err := s.queries.DeleteTutor(ctx, db.DeleteTutorParams{
		CourseParticipationID: tutorID,
		CoursePhaseID:         coursePhaseID,
	})

	if err != nil {
		log.Error("could not delete tutor: ", err)
		return errors.New("could not delete tutor")
	}
	return nil
}

// GetTutorAssignments lists every team each tutor of the phase leads, one entry per tutor and team.
func (s *TeamsService) GetTutorAssignments(ctx context.Context, coursePhaseID uuid.UUID) ([]teamDTO.TutorAssignment, error) {
	rows, err := s.queries.GetTutorAssignmentsByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		return nil, fmt.Errorf("could not get the tutors of course phase %s: %w", coursePhaseID, err)
	}
	return teamDTO.GetTutorAssignmentsFromDBModels(rows), nil
}

// GetTutorByUniversityLogin returns the tutor a university login belongs to, so another phase can
// tell which of a team's tutors the caller is.
func (s *TeamsService) GetTutorByUniversityLogin(ctx context.Context, coursePhaseID uuid.UUID, universityLogin string) (promptTypes.Person, error) {
	if universityLogin == "" {
		return promptTypes.Person{}, ErrTutorNotFound
	}

	tutor, err := s.queries.GetTutorByUniversityLogin(ctx, db.GetTutorByUniversityLoginParams{
		CoursePhaseID:   coursePhaseID,
		UniversityLogin: teamDTO.UniversityLoginParam(universityLogin),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return promptTypes.Person{}, ErrTutorNotFound
	}
	if err != nil {
		return promptTypes.Person{}, fmt.Errorf("failed to get tutor by university login: %w", err)
	}

	return promptTypes.Person{
		ID:        tutor.CourseParticipationID,
		FirstName: tutor.FirstName,
		LastName:  tutor.LastName,
	}, nil
}
