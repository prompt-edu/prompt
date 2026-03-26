package schemaModification

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas"
	"github.com/prompt-edu/prompt/servers/assessment/categories/categoryDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/utils"
	log "github.com/sirupsen/logrus"
)

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
func GetOrCopySchemaForWrite(
	ctx context.Context,
	queries db.Queries,
	schemaID uuid.UUID,
	entityID uuid.UUID,
	coursePhaseID uuid.UUID,
) (*PrepareSchemaModificationResult, error) {
	hasData, err := assessmentSchemas.CheckPhaseHasAssessmentData(ctx, coursePhaseID, schemaID)
	if err != nil {
		log.WithError(err).Error("Failed to check if phase has assessment data")
		return nil, errors.New("failed to check if phase has assessment data")
	}

	if hasData {
		log.Error("Modifications are not allowed on schemas with existing assessment data")
		return nil, errors.New("modifications are not allowed on schemas with existing assessment data")
	}

	isSchemaOwner, err := assessmentSchemas.CheckSchemaOwnership(ctx, schemaID, coursePhaseID)
	if err != nil {
		log.WithError(err).Error("Failed to check schema ownership")
		return nil, errors.New("failed to check schema ownership")
	}

	consumerPhases, err := assessmentSchemas.GetConsumerPhases(ctx, schemaID, coursePhaseID)
	if err != nil {
		log.WithError(err).Error("Failed to get consumer phases")
		return nil, errors.New("failed to get consumer phases")
	}
	hasConsumers := len(consumerPhases) > 0

	// SCENARIO 1: Phase owner modifying with consumers -> Copy schema for consumers with assessment data
	if isSchemaOwner && hasConsumers {
		// Copy schema for all consumers and update their assessment/evaluation references
		// This will automatically handle all categories and their competencies
		err = copySchemaForConsumersWithAssessmentData(ctx, queries, schemaID, consumerPhases)
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
		newSchemaID, err := copySchemaForConsumer(ctx, queries, schemaID, coursePhaseID)
		if err != nil {
			return nil, err
		}

		// If schema was copied, map the entity to the new schema
		if newSchemaID != schemaID && entityID != uuid.Nil {
			corresponding, err := queries.GetCorrespondingCompetencyInNewSchema(ctx, db.GetCorrespondingCompetencyInNewSchemaParams{
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
			categoryID, err := queries.GetCorrespondingCategoryInNewSchema(ctx, db.GetCorrespondingCategoryInNewSchemaParams{
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

func copySchemaForConsumer(ctx context.Context, queries db.Queries, oldSchemaID uuid.UUID, coursePhaseID uuid.UUID) (uuid.UUID, error) {
	courseIdentifier, err := utils.GetCourseIdentifierFromPhase(ctx, coursePhaseID)
	if err != nil {
		log.WithError(err).Error("Failed to get course identifier")
		return uuid.Nil, errors.New("failed to get course identifier")
	}

	copiedSchema, err := assessmentSchemas.CopyAssessmentSchema(ctx, coursePhaseID, oldSchemaID, courseIdentifier)
	if err != nil {
		log.WithError(err).Error("Failed to copy assessment schema")
		return uuid.Nil, errors.New("failed to copy assessment schema")
	}

	dbCategoriesWithCompetencies, err := queries.GetCategoriesWithCompetencies(ctx, oldSchemaID)
	if err != nil {
		log.WithError(err).Error("Failed to get categories with competencies from schema")
		return uuid.Nil, errors.New("failed to get categories with competencies")
	}
	categoriesWithCompetencies := categoryDTO.MapToCategoryWithCompetenciesDTO(dbCategoriesWithCompetencies)

	// Update assessment and evaluation competency references for ALL competencies in ALL categories
	for _, categoryWithCompetencies := range categoriesWithCompetencies {
		for _, competency := range categoryWithCompetencies.Competencies {
			newCompMapping, err := queries.GetCorrespondingCompetencyInNewSchema(ctx, db.GetCorrespondingCompetencyInNewSchemaParams{
				OldCompetencyID: competency.ID,
				NewSchemaID:     copiedSchema.ID,
			})
			if err != nil {
				log.WithError(err).WithField("competencyID", competency.ID).Warn("Failed to find corresponding competency in new schema, skipping")
				continue
			}

			// Update assessment and evaluation competency references
			err = assessmentSchemas.UpdateAssessmentAndEvaluationCompetencies(ctx, coursePhaseID, competency.ID, newCompMapping.CompetencyID)
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

func copySchemaForConsumersWithAssessmentData(ctx context.Context, queries db.Queries, oldSchemaID uuid.UUID, consumerPhases []uuid.UUID) error {
	for _, phaseID := range consumerPhases {
		hasAssessmentData, err := assessmentSchemas.CheckPhaseHasAssessmentData(ctx, phaseID, oldSchemaID)
		if err != nil {
			log.WithError(err).Error("Failed to check if phase has assessment data")
			return errors.New("failed to check if phase has assessment data")
		}
		if !hasAssessmentData {
			continue // No need to copy if no data exists
		}

		_, err = copySchemaForConsumer(ctx, queries, oldSchemaID, phaseID)
		if err != nil {
			log.WithError(err).Error("Failed to copy schema for consumer")
			return errors.New("failed to copy schema for consumer")
		}
	}

	return nil
}
