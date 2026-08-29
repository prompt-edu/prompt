package competencies

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas"
	"github.com/prompt-edu/prompt/servers/assessment/competencies/competencyDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/schemaModification"
	log "github.com/sirupsen/logrus"
)

type schemaProvider interface {
	CheckSchemaAccessibleForCoursePhase(ctx context.Context, coursePhaseID uuid.UUID, schemaID uuid.UUID) (bool, error)
}

type schemaModifier interface {
	GetOrCopySchemaForWrite(ctx context.Context, schemaID uuid.UUID, entityID uuid.UUID, coursePhaseID uuid.UUID) (*schemaModification.PrepareSchemaModificationResult, error)
}

type CompetencyService struct {
	queries            db.Queries
	conn               *pgxpool.Pool
	schemas            schemaProvider
	schemaModification schemaModifier
}

func NewCompetencyService(queries db.Queries, conn *pgxpool.Pool, schemas schemaProvider, schemaModification schemaModifier) *CompetencyService {
	return &CompetencyService{
		queries:            queries,
		conn:               conn,
		schemas:            schemas,
		schemaModification: schemaModification,
	}
}

func (s *CompetencyService) CreateCompetency(ctx context.Context, coursePhaseID uuid.UUID, req competencyDTO.CreateCompetencyRequest) error {
	category, err := s.queries.GetCategory(ctx, req.CategoryID)
	if err != nil {
		log.Error("could not get category: ", err)
		return errors.New("could not get category")
	}

	isAccessible, err := s.schemas.CheckSchemaAccessibleForCoursePhase(
		ctx,
		coursePhaseID,
		category.AssessmentSchemaID,
	)
	if err != nil {
		return err
	}
	if !isAccessible {
		return assessmentSchemas.ErrSchemaNotAccessible
	}

	result, err := s.schemaModification.GetOrCopySchemaForWrite(
		ctx,
		category.AssessmentSchemaID,
		req.CategoryID, // Pass category ID to get it mapped if schema is copied
		coursePhaseID,
	)
	if err != nil {
		return err
	}

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	err = qtx.CreateCompetency(ctx, db.CreateCompetencyParams{
		ID:                  uuid.New(),
		CategoryID:          result.TargetEntityID,
		Name:                req.Name,
		ShortName:           pgtype.Text{String: req.ShortName, Valid: true},
		Description:         pgtype.Text{String: req.Description, Valid: true},
		DescriptionVeryBad:  req.DescriptionVeryBad,
		DescriptionBad:      req.DescriptionBad,
		DescriptionOk:       req.DescriptionOk,
		DescriptionGood:     req.DescriptionGood,
		DescriptionVeryGood: req.DescriptionVeryGood,
		Weight:              req.Weight,
	})
	if err != nil {
		log.Error("could not create competency: ", err)
		return errors.New("could not create competency")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("could not commit competency creation: ", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *CompetencyService) GetCompetency(ctx context.Context, id uuid.UUID) (db.Competency, error) {
	competency, err := s.queries.GetCompetency(ctx, id)
	if err != nil {
		log.Error("could not get competency: ", err)
		return db.Competency{}, errors.New("could not get competency")
	}
	return competency, nil
}

func (s *CompetencyService) ListCompetencies(ctx context.Context) ([]db.Competency, error) {
	competencies, err := s.queries.ListCompetencies(ctx)
	if err != nil {
		log.Error("could not list competencies: ", err)
		return nil, errors.New("could not list competencies")
	}
	return competencies, nil
}

func (s *CompetencyService) ListCompetenciesForCoursePhase(ctx context.Context, coursePhaseID uuid.UUID) ([]db.Competency, error) {
	competencies, err := s.queries.ListCompetenciesForCoursePhase(
		ctx,
		pgtype.UUID{Bytes: coursePhaseID, Valid: true},
	)
	if err != nil {
		log.Error("could not list competencies: ", err)
		return nil, errors.New("could not list competencies")
	}
	return competencies, nil
}

func (s *CompetencyService) ListCompetenciesByCategory(ctx context.Context, categoryID uuid.UUID) ([]db.Competency, error) {
	competencies, err := s.queries.ListCompetenciesByCategory(ctx, categoryID)
	if err != nil {
		log.Error("could not list competencies by category: ", err)
		return nil, errors.New("could not list competencies by category")
	}
	return competencies, nil
}

func (s *CompetencyService) GetCompetencyForCoursePhase(ctx context.Context, coursePhaseID uuid.UUID, id uuid.UUID) (db.Competency, error) {
	schemaID, err := s.queries.GetAssessmentSchemaIDByCompetency(ctx, id)
	if err != nil {
		return db.Competency{}, err
	}

	isAccessible, err := s.schemas.CheckSchemaAccessibleForCoursePhase(ctx, coursePhaseID, schemaID)
	if err != nil {
		return db.Competency{}, err
	}
	if !isAccessible {
		return db.Competency{}, assessmentSchemas.ErrSchemaNotAccessible
	}

	return s.GetCompetency(ctx, id)
}

func (s *CompetencyService) ListCompetenciesByCategoryForCoursePhase(
	ctx context.Context,
	coursePhaseID uuid.UUID,
	categoryID uuid.UUID,
) ([]db.Competency, error) {
	category, err := s.queries.GetCategory(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	isAccessible, err := s.schemas.CheckSchemaAccessibleForCoursePhase(
		ctx,
		coursePhaseID,
		category.AssessmentSchemaID,
	)
	if err != nil {
		return nil, err
	}
	if !isAccessible {
		return nil, assessmentSchemas.ErrSchemaNotAccessible
	}

	return s.ListCompetenciesByCategory(ctx, categoryID)
}

func (s *CompetencyService) UpdateCompetency(ctx context.Context, id uuid.UUID, coursePhaseID uuid.UUID, req competencyDTO.UpdateCompetencyRequest) error {
	currentSchemaID, err := s.queries.GetAssessmentSchemaIDByCompetency(ctx, id)
	if err != nil {
		log.WithError(err).Error("Failed to get assessment schema ID for competency")
		return errors.New("failed to get assessment schema ID for competency")
	}

	isAccessible, err := s.schemas.CheckSchemaAccessibleForCoursePhase(ctx, coursePhaseID, currentSchemaID)
	if err != nil {
		return err
	}
	if !isAccessible {
		return assessmentSchemas.ErrSchemaNotAccessible
	}

	result, err := s.schemaModification.GetOrCopySchemaForWrite(
		ctx,
		currentSchemaID,
		id,
		coursePhaseID,
	)
	if err != nil {
		return err
	}

	err = s.queries.UpdateCompetency(ctx, db.UpdateCompetencyParams{
		ID:                  result.TargetEntityID,
		CategoryID:          req.CategoryID,
		Name:                req.Name,
		ShortName:           pgtype.Text{String: req.ShortName, Valid: true},
		Description:         pgtype.Text{String: req.Description, Valid: true},
		DescriptionVeryBad:  req.DescriptionVeryBad,
		DescriptionBad:      req.DescriptionBad,
		DescriptionOk:       req.DescriptionOk,
		DescriptionGood:     req.DescriptionGood,
		DescriptionVeryGood: req.DescriptionVeryGood,
		Weight:              req.Weight,
	})
	if err != nil {
		log.Error("could not update competency: ", err)
		return errors.New("could not update competency")
	}

	return nil
}

func (s *CompetencyService) DeleteCompetency(ctx context.Context, id uuid.UUID, coursePhaseID uuid.UUID) error {
	currentSchemaID, err := s.queries.GetAssessmentSchemaIDByCompetency(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		log.WithError(err).Error("Failed to get assessment schema ID for competency")
		return errors.New("failed to get assessment schema ID for competency")
	}

	isAccessible, err := s.schemas.CheckSchemaAccessibleForCoursePhase(ctx, coursePhaseID, currentSchemaID)
	if err != nil {
		return err
	}
	if !isAccessible {
		return assessmentSchemas.ErrSchemaNotAccessible
	}

	result, err := s.schemaModification.GetOrCopySchemaForWrite(
		ctx,
		currentSchemaID,
		id,
		coursePhaseID,
	)
	if err != nil {
		return err
	}

	err = s.queries.DeleteCompetency(ctx, result.TargetEntityID)
	if err != nil {
		log.Error("could not delete competency: ", err)
		return errors.New("could not delete competency")
	}

	return nil
}
