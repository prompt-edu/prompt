package categories

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/categories/categoryDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	log "github.com/sirupsen/logrus"
)

var (
	ErrIncompleteSchemaOrder   = errors.New("the order must list every category and competency of the schema exactly once")
	ErrDuplicateCompetencyName = errors.New("a category cannot contain two competencies with the same name")
)

// UpdateSchemaOrder sets the display order of the categories of a schema and of the competencies within them,
// moving competencies to the category they are listed under.
func (s *CategoryService) UpdateSchemaOrder(ctx context.Context, coursePhaseID uuid.UUID, req categoryDTO.UpdateSchemaOrderRequest) error {
	if len(req.Categories) == 0 {
		return ErrIncompleteSchemaOrder
	}

	// Like the other category writes, take the schema from the entity rather than the client: after a
	// copy-on-write the client may still hold the ID of the schema that was copied.
	firstCategory, err := s.queries.GetCategory(ctx, req.Categories[0].ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrIncompleteSchemaOrder
		}
		log.WithError(err).Error("Failed to get category")
		return errors.New("failed to get category")
	}
	schemaID := firstCategory.AssessmentSchemaID

	if err := s.ensureSchemaAccessible(ctx, coursePhaseID, schemaID); err != nil {
		return err
	}

	current, err := s.GetCategoriesWithCompetencies(ctx, schemaID)
	if err != nil {
		return err
	}
	if err := validateSchemaOrder(current, req.Categories); err != nil {
		return err
	}

	result, err := s.schemaModification.GetOrCopySchemaForWrite(ctx, schemaID, uuid.Nil, coursePhaseID)
	if err != nil {
		return err
	}

	order := req.Categories
	if result.TargetSchemaID != schemaID {
		order, err = s.mapSchemaOrderToCopy(ctx, req.Categories, result.TargetSchemaID)
		if err != nil {
			return err
		}
	}

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to begin transaction")
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := s.queries.WithTx(tx)

	for categoryIndex, category := range order {
		if err := qtx.UpdateCategorySortOrder(ctx, db.UpdateCategorySortOrderParams{
			ID:        category.ID,
			SortOrder: int32(categoryIndex),
		}); err != nil {
			log.WithError(err).Error("could not update category order")
			return errors.New("could not update category order")
		}

		for competencyIndex, competencyID := range category.CompetencyIDs {
			if err := qtx.UpdateCompetencyCategoryAndSortOrder(ctx, db.UpdateCompetencyCategoryAndSortOrderParams{
				ID:         competencyID,
				CategoryID: category.ID,
				SortOrder:  int32(competencyIndex),
			}); err != nil {
				log.WithError(err).Error("could not update competency order")
				return errors.New("could not update competency order")
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.WithError(err).Error("Failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// validateSchemaOrder checks that the order lists exactly the categories and competencies of the schema
// and that no category ends up with two competencies of the same name.
func validateSchemaOrder(current []categoryDTO.CategoryWithCompetencies, order []categoryDTO.CategoryOrder) error {
	categoryIDs := make(map[uuid.UUID]struct{}, len(current))
	competencyNames := make(map[uuid.UUID]string)
	for _, category := range current {
		categoryIDs[category.ID] = struct{}{}
		for _, competency := range category.Competencies {
			competencyNames[competency.ID] = competency.Name
		}
	}

	if len(order) != len(categoryIDs) {
		return ErrIncompleteSchemaOrder
	}

	seenCategories := make(map[uuid.UUID]struct{}, len(order))
	seenCompetencies := make(map[uuid.UUID]struct{}, len(competencyNames))
	for _, category := range order {
		if _, ok := categoryIDs[category.ID]; !ok {
			return ErrIncompleteSchemaOrder
		}
		if _, duplicate := seenCategories[category.ID]; duplicate {
			return ErrIncompleteSchemaOrder
		}
		seenCategories[category.ID] = struct{}{}

		namesInCategory := make(map[string]struct{}, len(category.CompetencyIDs))
		for _, competencyID := range category.CompetencyIDs {
			name, ok := competencyNames[competencyID]
			if !ok {
				return ErrIncompleteSchemaOrder
			}
			if _, duplicate := seenCompetencies[competencyID]; duplicate {
				return ErrIncompleteSchemaOrder
			}
			seenCompetencies[competencyID] = struct{}{}

			if _, duplicate := namesInCategory[name]; duplicate {
				return ErrDuplicateCompetencyName
			}
			namesInCategory[name] = struct{}{}
		}
	}

	if len(seenCompetencies) != len(competencyNames) {
		return ErrIncompleteSchemaOrder
	}

	return nil
}

// mapSchemaOrderToCopy translates an order given in the IDs of a schema into the IDs of its copy.
func (s *CategoryService) mapSchemaOrderToCopy(ctx context.Context, order []categoryDTO.CategoryOrder, copySchemaID uuid.UUID) ([]categoryDTO.CategoryOrder, error) {
	mapped := make([]categoryDTO.CategoryOrder, 0, len(order))
	for _, category := range order {
		categoryID, err := s.queries.GetCorrespondingCategoryInNewSchema(ctx, db.GetCorrespondingCategoryInNewSchemaParams{
			OldCategoryID: category.ID,
			NewSchemaID:   copySchemaID,
		})
		if err != nil {
			log.WithError(err).WithField("categoryID", category.ID).Error("Failed to find corresponding category in copied schema")
			return nil, errors.New("failed to find corresponding category in copied schema")
		}

		competencyIDs := make([]uuid.UUID, 0, len(category.CompetencyIDs))
		for _, competencyID := range category.CompetencyIDs {
			competency, err := s.queries.GetCorrespondingCompetencyInNewSchema(ctx, db.GetCorrespondingCompetencyInNewSchemaParams{
				OldCompetencyID: competencyID,
				NewSchemaID:     copySchemaID,
			})
			if err != nil {
				log.WithError(err).WithField("competencyID", competencyID).Error("Failed to find corresponding competency in copied schema")
				return nil, errors.New("failed to find corresponding competency in copied schema")
			}
			competencyIDs = append(competencyIDs, competency.CompetencyID)
		}

		mapped = append(mapped, categoryDTO.CategoryOrder{ID: categoryID, CompetencyIDs: competencyIDs})
	}

	return mapped, nil
}
