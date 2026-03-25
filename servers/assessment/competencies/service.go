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

type CompetencyService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var CompetencyServiceSingleton *CompetencyService

func CreateCompetency(ctx context.Context, coursePhaseID uuid.UUID, req competencyDTO.CreateCompetencyRequest) error {
	category, err := CompetencyServiceSingleton.queries.GetCategory(ctx, req.CategoryID)
	if err != nil {
		log.Error("could not get category: ", err)
		return errors.New("could not get category")
	}

	isAccessible, err := assessmentSchemas.CheckSchemaAccessibleForCoursePhase(
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

	result, err := schemaModification.GetOrCopySchemaForWrite(
		ctx,
		CompetencyServiceSingleton.queries,
		category.AssessmentSchemaID,
		req.CategoryID, // Pass category ID to get it mapped if schema is copied
		coursePhaseID,
	)
	if err != nil {
		return err
	}

	tx, err := CompetencyServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := CompetencyServiceSingleton.queries.WithTx(tx)

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

func GetCompetency(ctx context.Context, id uuid.UUID) (db.Competency, error) {
	competency, err := CompetencyServiceSingleton.queries.GetCompetency(ctx, id)
	if err != nil {
		log.Error("could not get competency: ", err)
		return db.Competency{}, errors.New("could not get competency")
	}
	return competency, nil
}

func ListCompetencies(ctx context.Context) ([]db.Competency, error) {
	competencies, err := CompetencyServiceSingleton.queries.ListCompetencies(ctx)
	if err != nil {
		log.Error("could not list competencies: ", err)
		return nil, errors.New("could not list competencies")
	}
	return competencies, nil
}

func ListCompetenciesForCoursePhase(ctx context.Context, coursePhaseID uuid.UUID) ([]db.Competency, error) {
	competencies, err := CompetencyServiceSingleton.queries.ListCompetencies(ctx)
	if err != nil {
		log.Error("could not list competencies: ", err)
		return nil, errors.New("could not list competencies")
	}

	filtered := make([]db.Competency, 0, len(competencies))
	for _, competency := range competencies {
		schemaID, err := CompetencyServiceSingleton.queries.GetAssessmentSchemaIDByCompetency(ctx, competency.ID)
		if err != nil {
			return nil, err
		}

		isAccessible, err := assessmentSchemas.CheckSchemaAccessibleForCoursePhase(ctx, coursePhaseID, schemaID)
		if err != nil {
			return nil, err
		}
		if isAccessible {
			filtered = append(filtered, competency)
		}
	}

	return filtered, nil
}

func ListCompetenciesByCategory(ctx context.Context, categoryID uuid.UUID) ([]db.Competency, error) {
	competencies, err := CompetencyServiceSingleton.queries.ListCompetenciesByCategory(ctx, categoryID)
	if err != nil {
		log.Error("could not list competencies by category: ", err)
		return nil, errors.New("could not list competencies by category")
	}
	return competencies, nil
}

func GetCompetencyForCoursePhase(ctx context.Context, coursePhaseID uuid.UUID, id uuid.UUID) (db.Competency, error) {
	schemaID, err := CompetencyServiceSingleton.queries.GetAssessmentSchemaIDByCompetency(ctx, id)
	if err != nil {
		return db.Competency{}, err
	}

	isAccessible, err := assessmentSchemas.CheckSchemaAccessibleForCoursePhase(ctx, coursePhaseID, schemaID)
	if err != nil {
		return db.Competency{}, err
	}
	if !isAccessible {
		return db.Competency{}, assessmentSchemas.ErrSchemaNotAccessible
	}

	return GetCompetency(ctx, id)
}

func ListCompetenciesByCategoryForCoursePhase(
	ctx context.Context,
	coursePhaseID uuid.UUID,
	categoryID uuid.UUID,
) ([]db.Competency, error) {
	category, err := CompetencyServiceSingleton.queries.GetCategory(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	isAccessible, err := assessmentSchemas.CheckSchemaAccessibleForCoursePhase(
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

	return ListCompetenciesByCategory(ctx, categoryID)
}

func UpdateCompetency(ctx context.Context, id uuid.UUID, coursePhaseID uuid.UUID, req competencyDTO.UpdateCompetencyRequest) error {
	currentSchemaID, err := CompetencyServiceSingleton.queries.GetAssessmentSchemaIDByCompetency(ctx, id)
	if err != nil {
		log.WithError(err).Error("Failed to get assessment schema ID for competency")
		return errors.New("failed to get assessment schema ID for competency")
	}

	isAccessible, err := assessmentSchemas.CheckSchemaAccessibleForCoursePhase(ctx, coursePhaseID, currentSchemaID)
	if err != nil {
		return err
	}
	if !isAccessible {
		return assessmentSchemas.ErrSchemaNotAccessible
	}

	result, err := schemaModification.GetOrCopySchemaForWrite(
		ctx,
		CompetencyServiceSingleton.queries,
		currentSchemaID,
		id,
		coursePhaseID,
	)
	if err != nil {
		return err
	}

	err = CompetencyServiceSingleton.queries.UpdateCompetency(ctx, db.UpdateCompetencyParams{
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

func DeleteCompetency(ctx context.Context, id uuid.UUID, coursePhaseID uuid.UUID) error {
	currentSchemaID, err := CompetencyServiceSingleton.queries.GetAssessmentSchemaIDByCompetency(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		log.WithError(err).Error("Failed to get assessment schema ID for competency")
		return errors.New("failed to get assessment schema ID for competency")
	}

	isAccessible, err := assessmentSchemas.CheckSchemaAccessibleForCoursePhase(ctx, coursePhaseID, currentSchemaID)
	if err != nil {
		return err
	}
	if !isAccessible {
		return assessmentSchemas.ErrSchemaNotAccessible
	}

	result, err := schemaModification.GetOrCopySchemaForWrite(
		ctx,
		CompetencyServiceSingleton.queries,
		currentSchemaID,
		id,
		coursePhaseID,
	)
	if err != nil {
		return err
	}

	err = CompetencyServiceSingleton.queries.DeleteCompetency(ctx, result.TargetEntityID)
	if err != nil {
		log.Error("could not delete competency: ", err)
		return errors.New("could not delete competency")
	}

	return nil
}
