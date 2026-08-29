package schemaModification

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas/assessmentSchemaDTO"
	"github.com/prompt-edu/prompt/servers/assessment/categories/categoryDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/utils"
	log "github.com/sirupsen/logrus"
)

type schemaProvider interface {
	CheckPhaseHasAssessmentData(ctx context.Context, phaseID uuid.UUID, schemaID uuid.UUID) (bool, error)
	CheckSchemaOwnership(ctx context.Context, schemaID uuid.UUID, coursePhaseID uuid.UUID) (bool, error)
	GetConsumerPhases(ctx context.Context, schemaID uuid.UUID, ownerPhaseID uuid.UUID) ([]uuid.UUID, error)
	CopyAssessmentSchema(ctx context.Context, coursePhaseID uuid.UUID, sourceSchemaID uuid.UUID, courseIdentifierPrefix string) (assessmentSchemaDTO.AssessmentSchema, error)
	UpdateCategoryAssessmentCategory(ctx context.Context, coursePhaseID uuid.UUID, oldCategoryID uuid.UUID, newCategoryID uuid.UUID) error
	UpdateAssessmentAndEvaluationCompetencies(ctx context.Context, coursePhaseID uuid.UUID, oldCompetencyID uuid.UUID, newCompetencyID uuid.UUID) error
}

type SchemaModificationService struct {
	schemas schemaProvider
	queries db.Queries
}

func NewSchemaModificationService(schemas schemaProvider, queries db.Queries) *SchemaModificationService {
	return &SchemaModificationService{
		schemas: schemas,
		queries: queries,
	}
}

type PrepareSchemaModificationResult struct {
	// TargetSchemaID is the schema ID to use for the operation (might be a copy)
	TargetSchemaID uuid.UUID
	// TargetEntityID is the updated entity ID if it was mapped to a new schema (for update/delete)
	TargetEntityID uuid.UUID
}

// GetOrCopySchemaForWrite handles all schema copying logic before any modification operation (category/competency).
// It determines if the phase owns the schema or is consuming a shared/global schema copy.
// For CREATE operations: pass schemaID and set entityID to uuid.Nil
// For UPDATE/DELETE operations: pass both schemaID and entityID
func (s *SchemaModificationService) GetOrCopySchemaForWrite(
	ctx context.Context,
	schemaID uuid.UUID,
	entityID uuid.UUID,
	coursePhaseID uuid.UUID,
) (*PrepareSchemaModificationResult, error) {
	hasData, err := s.schemas.CheckPhaseHasAssessmentData(ctx, coursePhaseID, schemaID)
	if err != nil {
		log.WithError(err).Error("Failed to check if phase has assessment data")
		return nil, errors.New("failed to check if phase has assessment data")
	}

	if hasData {
		log.Error("Modifications are not allowed on schemas with existing assessment data")
		return nil, errors.New("modifications are not allowed on schemas with existing assessment data")
	}

	isSchemaOwner, err := s.schemas.CheckSchemaOwnership(ctx, schemaID, coursePhaseID)
	if err != nil {
		log.WithError(err).Error("Failed to check schema ownership")
		return nil, errors.New("failed to check schema ownership")
	}

	consumerPhases, err := s.schemas.GetConsumerPhases(ctx, schemaID, coursePhaseID)
	if err != nil {
		log.WithError(err).Error("Failed to get consumer phases")
		return nil, errors.New("failed to get consumer phases")
	}
	hasConsumers := len(consumerPhases) > 0

	// SCENARIO 1: Phase owner modifying with consumers -> Copy schema for consumers with assessment data
	if isSchemaOwner && hasConsumers {
		// Copy schema for all consumers and update their assessment/evaluation references
		// This will automatically handle all categories and their competencies
		err = s.copySchemaForConsumersWithAssessmentData(ctx, schemaID, consumerPhases)
		if err != nil {
			return nil, err
		}

		return &PrepareSchemaModificationResult{
			TargetSchemaID: schemaID,
			TargetEntityID: entityID,
		}, nil
	}

	// SCENARIO 2: Consumer modifying shared/global schema -> Copy schema for this phase only
	if !isSchemaOwner {
		newSchemaID, err := s.copySchemaForConsumer(ctx, schemaID, coursePhaseID)
		if err != nil {
			return nil, err
		}

		// If schema was copied, map the entity to the new schema
		if newSchemaID != schemaID && entityID != uuid.Nil {
			corresponding, err := s.queries.GetCorrespondingCompetencyInNewSchema(ctx, db.GetCorrespondingCompetencyInNewSchemaParams{
				OldCompetencyID: entityID,
				NewSchemaID:     newSchemaID,
			})
			if err == nil {
				return &PrepareSchemaModificationResult{
					TargetSchemaID: newSchemaID,
					TargetEntityID: corresponding.CompetencyID,
				}, nil
			}

			// Try category mapping if competency mapping failed
			categoryID, err := s.queries.GetCorrespondingCategoryInNewSchema(ctx, db.GetCorrespondingCategoryInNewSchemaParams{
				OldCategoryID: entityID,
				NewSchemaID:   newSchemaID,
			})
			if err != nil {
				log.Error("Failed to find corresponding entity in new schema")
				return nil, errors.New("failed to find corresponding entity in new schema")
			}

			return &PrepareSchemaModificationResult{
				TargetSchemaID: newSchemaID,
				TargetEntityID: categoryID,
			}, nil
		}

		return &PrepareSchemaModificationResult{
			TargetSchemaID: newSchemaID,
			TargetEntityID: entityID,
		}, nil
	}

	// SCENARIO 3: No sharing concerns → Direct modification
	return &PrepareSchemaModificationResult{
		TargetSchemaID: schemaID,
		TargetEntityID: entityID,
	}, nil
}

func (s *SchemaModificationService) copySchemaForConsumer(ctx context.Context, oldSchemaID uuid.UUID, coursePhaseID uuid.UUID) (uuid.UUID, error) {
	courseIdentifier, err := utils.GetCourseIdentifierFromPhase(ctx, coursePhaseID)
	if err != nil {
		log.WithError(err).Error("Failed to get course identifier")
		return uuid.Nil, errors.New("failed to get course identifier")
	}

	copiedSchema, err := s.schemas.CopyAssessmentSchema(ctx, coursePhaseID, oldSchemaID, courseIdentifier)
	if err != nil {
		log.WithError(err).Error("Failed to copy assessment schema")
		return uuid.Nil, errors.New("failed to copy assessment schema")
	}

	dbCategoriesWithCompetencies, err := s.queries.GetCategoriesWithCompetencies(ctx, oldSchemaID)
	if err != nil {
		log.WithError(err).Error("Failed to get categories with competencies from schema")
		return uuid.Nil, errors.New("failed to get categories with competencies")
	}
	categoriesWithCompetencies := categoryDTO.MapToCategoryWithCompetenciesDTO(dbCategoriesWithCompetencies)

	// Update assessment and evaluation competency references for ALL competencies in ALL categories,
	// and remap per-category student comments (category_assessment) to the new category IDs.
	for _, categoryWithCompetencies := range categoriesWithCompetencies {
		newCategoryID, err := s.queries.GetCorrespondingCategoryInNewSchema(ctx, db.GetCorrespondingCategoryInNewSchemaParams{
			OldCategoryID: categoryWithCompetencies.ID,
			NewSchemaID:   copiedSchema.ID,
		})
		if err == nil {
			if err := s.schemas.UpdateCategoryAssessmentCategory(ctx, coursePhaseID, categoryWithCompetencies.ID, newCategoryID); err != nil {
				return uuid.Nil, errors.New("failed to update category_assessment categories")
			}
		} else {
			log.WithError(err).WithField("categoryID", categoryWithCompetencies.ID).Warn("Failed to find corresponding category in new schema, category_assessment rows will be left untouched")
		}

		for _, competency := range categoryWithCompetencies.Competencies {
			newCompMapping, err := s.queries.GetCorrespondingCompetencyInNewSchema(ctx, db.GetCorrespondingCompetencyInNewSchemaParams{
				OldCompetencyID: competency.ID,
				NewSchemaID:     copiedSchema.ID,
			})
			if err != nil {
				log.WithError(err).WithField("competencyID", competency.ID).Warn("Failed to find corresponding competency in new schema, skipping")
				continue
			}

			// Update assessment and evaluation competency references
			err = s.schemas.UpdateAssessmentAndEvaluationCompetencies(ctx, coursePhaseID, competency.ID, newCompMapping.CompetencyID)
			if err != nil {
				log.WithError(err).Error("Failed to update competency references")
				return uuid.Nil, errors.New("failed to update competency references")
			}
		}
	}

	err = coursePhaseConfig.UpdateCoursePhaseConfigAssessmentSchema(ctx, coursePhaseID, oldSchemaID, copiedSchema.ID)
	if err != nil {
		log.Error("Failed to update course phase config with new schema")
		return uuid.Nil, errors.New("failed to update course phase config")
	}

	return copiedSchema.ID, nil
}

func (s *SchemaModificationService) copySchemaForConsumersWithAssessmentData(ctx context.Context, oldSchemaID uuid.UUID, consumerPhases []uuid.UUID) error {
	for _, phaseID := range consumerPhases {
		hasAssessmentData, err := s.schemas.CheckPhaseHasAssessmentData(ctx, phaseID, oldSchemaID)
		if err != nil {
			log.WithError(err).Error("Failed to check if phase has assessment data")
			return errors.New("failed to check if phase has assessment data")
		}
		if !hasAssessmentData {
			continue // No need to copy if no data exists
		}

		_, err = s.copySchemaForConsumer(ctx, oldSchemaID, phaseID)
		if err != nil {
			log.WithError(err).Error("Failed to copy schema for consumer")
			return errors.New("failed to copy schema for consumer")
		}
	}

	return nil
}
