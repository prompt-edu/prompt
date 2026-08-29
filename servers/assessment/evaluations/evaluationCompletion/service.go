package evaluationCompletion

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationCompletion/evaluationCompletionDTO"
	log "github.com/sirupsen/logrus"
)

type teamsResolver func(ctx context.Context, authHeader string, coursePhaseID uuid.UUID) ([]promptTypes.Team, error)

type coursePhaseConfigProvider interface {
	IsSelfEvaluationDeadlinePassed(ctx context.Context, coursePhaseID uuid.UUID) (bool, error)
	IsPeerEvaluationDeadlinePassed(ctx context.Context, coursePhaseID uuid.UUID) (bool, error)
	IsTutorEvaluationDeadlinePassed(ctx context.Context, coursePhaseID uuid.UUID) (bool, error)
}

type EvaluationCompletionService struct {
	queries                db.Queries
	conn                   *pgxpool.Pool
	getTeamsForCoursePhase teamsResolver
	coursePhaseConfig      coursePhaseConfigProvider
}

func NewEvaluationCompletionService(queries db.Queries, conn *pgxpool.Pool, getTeamsForCoursePhase teamsResolver, coursePhaseConfig coursePhaseConfigProvider) *EvaluationCompletionService {
	return &EvaluationCompletionService{
		queries:                queries,
		conn:                   conn,
		getTeamsForCoursePhase: getTeamsForCoursePhase,
		coursePhaseConfig:      coursePhaseConfig,
	}
}

var ErrInvalidEvaluationType = errors.New("invalid evaluation type")
var ErrSelfEvaluationTargetMismatch = errors.New("self evaluation must target its author")
var ErrPeerEvaluationTargetNotInTeam = errors.New("peer evaluation must target a member of the author's team")
var ErrTutorEvaluationTargetNotTeamTutor = errors.New("tutor evaluation must target a tutor of the author's team")
var ErrAuthorHasNoTeam = errors.New("author is not a member of any team in this course phase")
var ErrEvaluationAlreadyCompleted = errors.New("evaluation completion already exists and is marked as completed")

// IsTargetAuthorizationError reports whether err means the caller may not write a record
// about the requested subject, so routers can answer 403 instead of 500.
func IsTargetAuthorizationError(err error) bool {
	return errors.Is(err, ErrPeerEvaluationTargetNotInTeam) ||
		errors.Is(err, ErrTutorEvaluationTargetNotTeamTutor) ||
		errors.Is(err, ErrAuthorHasNoTeam)
}

func (s *EvaluationCompletionService) CheckEvaluationIsEditable(ctx context.Context, qtx *db.Queries, authHeader string, courseParticipationID, coursePhaseID, authorCourseParticipationID uuid.UUID, evaluationType assessmentType.AssessmentType) error {
	if err := s.checkEvaluationTarget(ctx, authHeader, coursePhaseID, courseParticipationID, authorCourseParticipationID, evaluationType); err != nil {
		return err
	}

	var open bool
	var err error
	switch evaluationType {
	case assessmentType.Self:
		open, err = qtx.IsSelfEvaluationOpen(ctx, coursePhaseID)
	case assessmentType.Peer:
		open, err = qtx.IsPeerEvaluationOpen(ctx, coursePhaseID)
	case assessmentType.Tutor:
		open, err = qtx.IsTutorEvaluationOpen(ctx, coursePhaseID)
	default:
		return ErrInvalidEvaluationType
	}
	if err != nil {
		log.Error("could not check if evaluation is open: ", err)
		return errors.New("could not check if evaluation is open")
	}
	if !open {
		return coursePhaseConfig.ErrNotStarted
	}

	exists, err := qtx.CheckEvaluationCompletionExists(ctx, db.CheckEvaluationCompletionExistsParams{
		CourseParticipationID:       courseParticipationID,
		CoursePhaseID:               coursePhaseID,
		AuthorCourseParticipationID: authorCourseParticipationID,
	})
	if err != nil {
		log.Error("could not check evaluation completion existence: ", err)
		return errors.New("could not check evaluation completion existence")
	}
	if exists {
		completion, err := qtx.GetEvaluationCompletion(ctx, db.GetEvaluationCompletionParams{
			CourseParticipationID:       courseParticipationID,
			CoursePhaseID:               coursePhaseID,
			AuthorCourseParticipationID: authorCourseParticipationID,
		})
		if err != nil {
			log.Error("could not get evaluation completion: ", err)
			return errors.New("could not get evaluation completion")
		}

		if completion.Completed {
			return ErrEvaluationAlreadyCompleted
		}
	}
	return nil
}

// checkEvaluationTarget enforces who a record of the given type may be written about.
// A self evaluation must target its author; a peer evaluation must target a member of the
// author's team; a tutor evaluation must target one of that team's tutors.
//
// The default branch returns nil rather than rejecting: unknown types are the responsibility of
// the evaluation-type switch in CheckEvaluationIsEditable, which answers ErrInvalidEvaluationType.
func (s *EvaluationCompletionService) checkEvaluationTarget(ctx context.Context, authHeader string, coursePhaseID, courseParticipationID, authorCourseParticipationID uuid.UUID, evaluationType assessmentType.AssessmentType) error {
	switch evaluationType {
	case assessmentType.Self:
		if courseParticipationID != authorCourseParticipationID {
			return ErrSelfEvaluationTargetMismatch
		}
		return nil
	case assessmentType.Peer, assessmentType.Tutor:
		return s.checkTeamTarget(ctx, authHeader, coursePhaseID, courseParticipationID, authorCourseParticipationID, evaluationType)
	default:
		return nil
	}
}

func (s *EvaluationCompletionService) checkTeamTarget(ctx context.Context, authHeader string, coursePhaseID, courseParticipationID, authorCourseParticipationID uuid.UUID, evaluationType assessmentType.AssessmentType) error {
	teams, err := s.getTeamsForCoursePhase(ctx, authHeader, coursePhaseID)
	if err != nil {
		log.Error("could not fetch teams to validate the evaluation target: ", err)
		return errors.New("could not fetch teams to validate the evaluation target")
	}

	team, found := teamOfMember(teams, authorCourseParticipationID)
	if !found {
		return ErrAuthorHasNoTeam
	}

	if evaluationType == assessmentType.Tutor {
		if !containsPerson(team.Tutors, courseParticipationID) {
			return ErrTutorEvaluationTargetNotTeamTutor
		}
		return nil
	}

	if courseParticipationID == authorCourseParticipationID || !containsPerson(team.Members, courseParticipationID) {
		return ErrPeerEvaluationTargetNotInTeam
	}
	return nil
}

// teamOfMember returns the first team the participant belongs to. A participant belongs to exactly
// one team per course phase, so the first match is the only match.
func teamOfMember(teams []promptTypes.Team, courseParticipationID uuid.UUID) (promptTypes.Team, bool) {
	for _, team := range teams {
		if containsPerson(team.Members, courseParticipationID) {
			return team, true
		}
	}
	return promptTypes.Team{}, false
}

func containsPerson(persons []promptTypes.Person, courseParticipationID uuid.UUID) bool {
	for _, person := range persons {
		if person.ID == courseParticipationID {
			return true
		}
	}
	return false
}

func (s *EvaluationCompletionService) CreateOrUpdateEvaluationCompletion(ctx context.Context, authHeader string, req evaluationCompletionDTO.EvaluationCompletion) error {
	err := s.CheckEvaluationIsEditable(ctx, &s.queries, authHeader, req.CourseParticipationID, req.CoursePhaseID, req.AuthorCourseParticipationID, req.Type)
	if err != nil {
		return err
	}

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := s.queries.WithTx(tx)

	err = qtx.CreateOrUpdateEvaluationCompletion(ctx, db.CreateOrUpdateEvaluationCompletionParams{
		CourseParticipationID:       req.CourseParticipationID,
		CoursePhaseID:               req.CoursePhaseID,
		AuthorCourseParticipationID: req.AuthorCourseParticipationID,
		CompletedAt:                 pgtype.Timestamptz{Time: time.Now(), Valid: true},
		Completed:                   req.Completed,
	})
	if err != nil {
		log.Error("could not create or update evaluation completion: ", err)
		return errors.New("could not create or update evaluation completion")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("could not commit evaluation completion: ", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *EvaluationCompletionService) MarkEvaluationAsCompleted(ctx context.Context, authHeader string, req evaluationCompletionDTO.EvaluationCompletion) error {
	err := s.CheckEvaluationIsEditable(ctx, &s.queries, authHeader, req.CourseParticipationID, req.CoursePhaseID, req.AuthorCourseParticipationID, req.Type)
	if err != nil {
		return err
	}

	// Check if there are remaining evaluations before marking as completed
	remainingEvaluations, err := s.queries.CountRemainingEvaluationsForStudent(ctx, db.CountRemainingEvaluationsForStudentParams{
		CourseParticipationID:       req.CourseParticipationID,
		AuthorCourseParticipationID: req.AuthorCourseParticipationID,
		CoursePhaseID:               req.CoursePhaseID,
		Column4:                     assessmentType.MapDTOtoDBAssessmentType(req.Type),
	})
	if err != nil {
		log.Error("could not check remaining evaluations: ", err)
		return errors.New("could not check remaining evaluations")
	}

	if remainingEvaluations > 0 {
		log.Warnf("cannot mark evaluation as completed: %d evaluations still remaining", remainingEvaluations)
		return fmt.Errorf("cannot mark evaluation as completed: %d evaluations still remaining", remainingEvaluations)
	}

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := s.queries.WithTx(tx)

	err = qtx.MarkEvaluationAsFinished(ctx, db.MarkEvaluationAsFinishedParams{
		CourseParticipationID:       req.CourseParticipationID,
		CoursePhaseID:               req.CoursePhaseID,
		AuthorCourseParticipationID: req.AuthorCourseParticipationID,
		CompletedAt:                 pgtype.Timestamptz{Time: time.Now(), Valid: true},
		Type:                        assessmentType.MapDTOtoDBAssessmentType(req.Type),
	})
	if err != nil {
		log.Error("could not mark evaluation as finished: ", err)
		return errors.New("could not mark evaluation as finished")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("could not commit evaluation completion: ", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *EvaluationCompletionService) UnmarkEvaluationAsCompleted(ctx context.Context, courseParticipationID, coursePhaseID, authorCourseParticipationID uuid.UUID) error {
	// Get the evaluation completion to determine its type
	dbCompletion, err := s.queries.GetEvaluationCompletion(ctx, db.GetEvaluationCompletionParams{
		CourseParticipationID:       courseParticipationID,
		CoursePhaseID:               coursePhaseID,
		AuthorCourseParticipationID: authorCourseParticipationID,
	})
	if err != nil {
		log.Error("could not get evaluation completion: ", err)
		return errors.New("could not get evaluation completion")
	}

	completion := evaluationCompletionDTO.MapDBEvaluationCompletionToEvaluationCompletionDTO(dbCompletion)
	switch completion.Type {
	case assessmentType.Self:
		deadlinePassed, err := s.coursePhaseConfig.IsSelfEvaluationDeadlinePassed(ctx, coursePhaseID)
		if err != nil {
			return err
		}
		if deadlinePassed {
			return coursePhaseConfig.ErrDeadlinePassed
		}
	case assessmentType.Peer:
		deadlinePassed, err := s.coursePhaseConfig.IsPeerEvaluationDeadlinePassed(ctx, coursePhaseID)
		if err != nil {
			return err
		}
		if deadlinePassed {
			return coursePhaseConfig.ErrDeadlinePassed
		}
	case assessmentType.Tutor:
		deadlinePassed, err := s.coursePhaseConfig.IsTutorEvaluationDeadlinePassed(ctx, coursePhaseID)
		if err != nil {
			return err
		}
		if deadlinePassed {
			return coursePhaseConfig.ErrDeadlinePassed
		}
	}

	err = s.queries.UnmarkEvaluationAsFinished(ctx, db.UnmarkEvaluationAsFinishedParams{
		CourseParticipationID:       courseParticipationID,
		CoursePhaseID:               coursePhaseID,
		AuthorCourseParticipationID: authorCourseParticipationID,
	})
	if err != nil {
		log.Error("could not unmark evaluation as finished: ", err)
		return errors.New("could not unmark evaluation as finished")
	}
	return nil
}

func (s *EvaluationCompletionService) DeleteEvaluationCompletion(ctx context.Context, courseParticipationID, coursePhaseID, authorCourseParticipationID uuid.UUID) error {
	err := s.queries.DeleteEvaluationCompletion(ctx, db.DeleteEvaluationCompletionParams{
		CourseParticipationID:       courseParticipationID,
		CoursePhaseID:               coursePhaseID,
		AuthorCourseParticipationID: authorCourseParticipationID,
	})
	if err != nil {
		log.Error("could not delete evaluation completion: ", err)
		return errors.New("could not delete evaluation completion")
	}
	return nil
}

func (s *EvaluationCompletionService) ListEvaluationCompletionsByCoursePhase(ctx context.Context, coursePhaseID uuid.UUID) ([]db.EvaluationCompletion, error) {
	completions, err := s.queries.GetEvaluationCompletionsByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not get evaluation completions by course phase: ", err)
		return nil, errors.New("could not get evaluation completions by course phase")
	}
	return completions, nil
}

func (s *EvaluationCompletionService) GetEvaluationCompletionForParticipantInPhase(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) ([]db.EvaluationCompletion, error) {
	completions, err := s.queries.GetEvaluationCompletionsForParticipantInPhase(ctx, db.GetEvaluationCompletionsForParticipantInPhaseParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("could not get evaluation completions for participant in phase: ", err)
		return nil, errors.New("could not get evaluation completions for participant in phase")
	}
	return completions, nil
}

func (s *EvaluationCompletionService) GetEvaluationCompletionsForAuthorInPhase(ctx context.Context, authorCourseParticipationID, coursePhaseID uuid.UUID) ([]db.EvaluationCompletion, error) {
	completions, err := s.queries.GetEvaluationCompletionsForAuthorInPhase(ctx, db.GetEvaluationCompletionsForAuthorInPhaseParams{
		AuthorCourseParticipationID: authorCourseParticipationID,
		CoursePhaseID:               coursePhaseID,
	})
	if err != nil {
		log.Error("could not get evaluation completions for author in phase: ", err)
		return nil, errors.New("could not get evaluation completions for author in phase")
	}
	return completions, nil
}
