package categories

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
	"github.com/prompt-edu/prompt/servers/assessment/categories/categoryDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/schemaModification"
	log "github.com/sirupsen/logrus"
)

type CategoryService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var CategoryServiceSingleton *CategoryService

func ensureSchemaAccessible(ctx context.Context, coursePhaseID uuid.UUID, schemaID uuid.UUID) error {
	isAccessible, err := assessmentSchemas.CheckSchemaAccessibleForCoursePhase(ctx, coursePhaseID, schemaID)
	if err != nil {
		return err
	}
	if !isAccessible {
		return assessmentSchemas.ErrSchemaNotAccessible
	}

	return nil
}

func CreateCategory(ctx context.Context, coursePhaseID uuid.UUID, req categoryDTO.CreateCategoryRequest) error {
	if err := ensureSchemaAccessible(ctx, coursePhaseID, req.AssessmentSchemaID); err != nil {
		return err
	}

	result, err := schemaModification.GetOrCopySchemaForWrite(
		ctx,
		CategoryServiceSingleton.queries,
		req.AssessmentSchemaID,
		uuid.Nil, // No entity ID for create operations
		coursePhaseID,
	)
	if err != nil {
		return err
	}

	tx, err := CategoryServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := CategoryServiceSingleton.queries.WithTx(tx)

	err = qtx.CreateCategory(ctx, db.CreateCategoryParams{
		ID:                 uuid.New(),
		Name:               req.Name,
		ShortName:          pgtype.Text{String: req.ShortName, Valid: true},
		Description:        pgtype.Text{String: req.Description, Valid: true},
		Weight:             req.Weight,
		AssessmentSchemaID: result.TargetSchemaID,
	})
	if err != nil {
		log.Error("could not create category: ", err)
		return errors.New("could not create category")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func GetCategory(ctx context.Context, id uuid.UUID) (db.Category, error) {
	category, err := CategoryServiceSingleton.queries.GetCategory(ctx, id)
	if err != nil {
		log.Error("could not get category: ", err)
		return db.Category{}, errors.New("could not get category")
	}
	return category, nil
}

func ListCategories(ctx context.Context) ([]db.Category, error) {
	categories, err := CategoryServiceSingleton.queries.ListCategories(ctx)
	if err != nil {
		log.Error("could not list categories: ", err)
		return nil, errors.New("could not list categories")
	}
	return categories, nil
}

func ListCategoriesForCoursePhase(ctx context.Context, coursePhaseID uuid.UUID) ([]db.Category, error) {
	categories, err := CategoryServiceSingleton.queries.ListCategories(ctx)
	if err != nil {
		log.Error("could not list categories: ", err)
		return nil, errors.New("could not list categories")
	}

	accessibleSchemas, err := assessmentSchemas.ListAssessmentSchemasForCoursePhase(ctx, coursePhaseID)
	if err != nil {
		return nil, err
	}

	accessibleSchemaIDs := make(map[uuid.UUID]struct{}, len(accessibleSchemas))
	for _, schema := range accessibleSchemas {
		accessibleSchemaIDs[schema.ID] = struct{}{}
	}

	filtered := make([]db.Category, 0, len(categories))
	for _, category := range categories {
		if _, ok := accessibleSchemaIDs[category.AssessmentSchemaID]; ok {
			filtered = append(filtered, category)
		}
	}

	return filtered, nil
}

func UpdateCategory(ctx context.Context, id uuid.UUID, coursePhaseID uuid.UUID, req categoryDTO.UpdateCategoryRequest) error {
	currentCategory, err := CategoryServiceSingleton.queries.GetCategory(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		log.WithError(err).Error("Failed to get current category")
		return errors.New("failed to get current category")
	}
	currentSchemaID := currentCategory.AssessmentSchemaID

	if err := ensureSchemaAccessible(ctx, coursePhaseID, currentSchemaID); err != nil {
		return err
	}

	result, err := schemaModification.GetOrCopySchemaForWrite(
		ctx,
		CategoryServiceSingleton.queries,
		currentSchemaID,
		id,
		coursePhaseID,
	)
	if err != nil {
		return err
	}

	tx, err := CategoryServiceSingleton.conn.Begin(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to begin transaction")
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := CategoryServiceSingleton.queries.WithTx(tx)

	err = qtx.UpdateCategory(ctx, db.UpdateCategoryParams{
		ID:                 result.TargetEntityID,
		Name:               req.Name,
		ShortName:          pgtype.Text{String: req.ShortName, Valid: true},
		Description:        pgtype.Text{String: req.Description, Valid: true},
		Weight:             req.Weight,
		AssessmentSchemaID: result.TargetSchemaID,
	})
	if err != nil {
		log.Error("could not update category: ", err)
		return errors.New("could not update category")
	}

	if err := tx.Commit(ctx); err != nil {
		log.WithError(err).Error("Failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func DeleteCategory(ctx context.Context, id uuid.UUID, coursePhaseID uuid.UUID) error {
	currentCategory, err := CategoryServiceSingleton.queries.GetCategory(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		log.Error("could not get category: ", err)
		return errors.New("could not get category")
	}
	currentSchemaID := currentCategory.AssessmentSchemaID

	if err := ensureSchemaAccessible(ctx, coursePhaseID, currentSchemaID); err != nil {
		return err
	}

	result, err := schemaModification.GetOrCopySchemaForWrite(
		ctx,
		CategoryServiceSingleton.queries,
		currentSchemaID,
		id,
		coursePhaseID,
	)
	if err != nil {
		return err
	}

	tx, err := CategoryServiceSingleton.conn.Begin(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to begin transaction")
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := CategoryServiceSingleton.queries.WithTx(tx)

	err = qtx.DeleteCategory(ctx, result.TargetEntityID)
	if err != nil {
		log.Error("could not delete category: ", err)
		return errors.New("could not delete category")
	}

	if err := tx.Commit(ctx); err != nil {
		log.WithError(err).Error("Failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func GetCategoriesWithCompetencies(ctx context.Context, assessmentSchemaID uuid.UUID) ([]categoryDTO.CategoryWithCompetencies, error) {
	dbRows, err := CategoryServiceSingleton.queries.GetCategoriesWithCompetencies(ctx, assessmentSchemaID)
	if err != nil {
		log.Error("could not get categories with competencies: ", err)
		return nil, errors.New("could not get categories with competencies")
	}
	return categoryDTO.MapToCategoryWithCompetenciesDTO(dbRows), nil
}
