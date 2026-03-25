package assessmentSchemas

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas/assessmentSchemaDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	log "github.com/sirupsen/logrus"
)

type AssessmentSchemaService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var AssessmentSchemaServiceSingleton *AssessmentSchemaService
var ErrSchemaNotAccessible = errors.New("assessment schema is not accessible for this course phase")

func NewAssessmentSchemaService(queries db.Queries, conn *pgxpool.Pool) *AssessmentSchemaService {
	return &AssessmentSchemaService{
		queries: queries,
		conn:    conn,
	}
}

func CreateAssessmentSchema(ctx context.Context, req assessmentSchemaDTO.CreateAssessmentSchemaRequest) (assessmentSchemaDTO.AssessmentSchema, error) {
	return insertAssessmentSchema(ctx, pgtype.UUID{Valid: false}, req)
}

func CreateAssessmentSchemaForCoursePhase(
	ctx context.Context,
	coursePhaseID uuid.UUID,
	req assessmentSchemaDTO.CreateAssessmentSchemaRequest,
) (assessmentSchemaDTO.AssessmentSchema, error) {
	return insertAssessmentSchema(ctx, pgtype.UUID{Bytes: coursePhaseID, Valid: true}, req)
}

func insertAssessmentSchema(
	ctx context.Context,
	sourcePhaseID pgtype.UUID,
	req assessmentSchemaDTO.CreateAssessmentSchemaRequest,
) (assessmentSchemaDTO.AssessmentSchema, error) {
	tx, err := AssessmentSchemaServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return assessmentSchemaDTO.AssessmentSchema{}, err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := AssessmentSchemaServiceSingleton.queries.WithTx(tx)

	schemaID := uuid.New()

	var description pgtype.Text
	if req.Description != "" {
		description = pgtype.Text{String: req.Description, Valid: true}
	}

	err = qtx.CreateAssessmentSchema(ctx, db.CreateAssessmentSchemaParams{
		ID:            schemaID,
		Name:          req.Name,
		Description:   description,
		SourcePhaseID: sourcePhaseID,
	})
	if err != nil {
		log.WithError(err).Error("Failed to create assessment schema")
		return assessmentSchemaDTO.AssessmentSchema{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return assessmentSchemaDTO.AssessmentSchema{}, err
	}

	// Fetch the created schema
	schema, err := GetAssessmentSchema(ctx, schemaID)
	if err != nil {
		return assessmentSchemaDTO.AssessmentSchema{}, err
	}

	return schema, nil
}

func GetAssessmentSchema(ctx context.Context, schemaID uuid.UUID) (assessmentSchemaDTO.AssessmentSchema, error) {
	schema, err := AssessmentSchemaServiceSingleton.queries.GetAssessmentSchema(ctx, schemaID)
	if err != nil {
		log.WithError(err).Error("Failed to get assessment schema")
		return assessmentSchemaDTO.AssessmentSchema{}, err
	}

	return assessmentSchemaDTO.MapDBAssessmentSchemaToDTOAssessmentSchema(schema), nil
}

func ListAssessmentSchemas(ctx context.Context) ([]assessmentSchemaDTO.AssessmentSchema, error) {
	schemas, err := AssessmentSchemaServiceSingleton.queries.ListAssessmentSchemas(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to list assessment schemas")
		return nil, err
	}

	result := make([]assessmentSchemaDTO.AssessmentSchema, len(schemas))
	for i, schema := range schemas {
		result[i] = assessmentSchemaDTO.MapDBAssessmentSchemaToDTOAssessmentSchema(schema)
	}

	return result, nil
}

func ListAssessmentSchemasForCoursePhase(
	ctx context.Context,
	coursePhaseID uuid.UUID,
) ([]assessmentSchemaDTO.AssessmentSchema, error) {
	schemas, err := AssessmentSchemaServiceSingleton.queries.ListAssessmentSchemasForCoursePhase(
		ctx,
		pgtype.UUID{Bytes: coursePhaseID, Valid: true},
	)
	if err != nil {
		log.WithError(err).Error("Failed to list accessible assessment schemas")
		return nil, err
	}

	result := make([]assessmentSchemaDTO.AssessmentSchema, len(schemas))
	for i, schema := range schemas {
		result[i] = assessmentSchemaDTO.MapDBAssessmentSchemaToDTOAssessmentSchema(schema)
	}

	return result, nil
}

func CheckSchemaAccessibleForCoursePhase(
	ctx context.Context,
	coursePhaseID uuid.UUID,
	schemaID uuid.UUID,
) (bool, error) {
	isAccessible, err := AssessmentSchemaServiceSingleton.queries.CheckSchemaAccessibleForCoursePhase(
		ctx,
		db.CheckSchemaAccessibleForCoursePhaseParams{
			CoursePhaseID: pgtype.UUID{Bytes: coursePhaseID, Valid: true},
			SchemaID:      schemaID,
		},
	)
	if err != nil {
		log.WithError(err).Error("Failed to check schema accessibility")
		return false, err
	}

	return isAccessible, nil
}

func GetAssessmentSchemaForCoursePhase(
	ctx context.Context,
	coursePhaseID uuid.UUID,
	schemaID uuid.UUID,
) (assessmentSchemaDTO.AssessmentSchema, error) {
	isAccessible, err := CheckSchemaAccessibleForCoursePhase(ctx, coursePhaseID, schemaID)
	if err != nil {
		return assessmentSchemaDTO.AssessmentSchema{}, err
	}
	if !isAccessible {
		return assessmentSchemaDTO.AssessmentSchema{}, ErrSchemaNotAccessible
	}

	return GetAssessmentSchema(ctx, schemaID)
}

func UpdateAssessmentSchema(ctx context.Context, schemaID uuid.UUID, req assessmentSchemaDTO.UpdateAssessmentSchemaRequest) error {
	// First check if schema exists
	_, err := GetAssessmentSchema(ctx, schemaID)
	if err != nil {
		return err
	}

	tx, err := AssessmentSchemaServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := AssessmentSchemaServiceSingleton.queries.WithTx(tx)

	var description pgtype.Text
	if req.Description != "" {
		description = pgtype.Text{String: req.Description, Valid: true}
	}

	err = qtx.UpdateAssessmentSchema(ctx, db.UpdateAssessmentSchemaParams{
		ID:          schemaID,
		Name:        req.Name,
		Description: description,
	})
	if err != nil {
		log.WithError(err).Error("Failed to update assessment schema")
		return err
	}

	return tx.Commit(ctx)
}

func DeleteAssessmentSchema(ctx context.Context, schemaID uuid.UUID) error {
	tx, err := AssessmentSchemaServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := AssessmentSchemaServiceSingleton.queries.WithTx(tx)

	err = qtx.DeleteAssessmentSchema(ctx, schemaID)
	if err != nil {
		log.WithError(err).Error("Failed to delete assessment schema")
		return err
	}

	return tx.Commit(ctx)
}

func GetCoursePhasesByAssessmentSchema(ctx context.Context, assessmentSchemaID uuid.UUID) ([]uuid.UUID, error) {
	coursePhaseIDs, err := AssessmentSchemaServiceSingleton.queries.GetCoursePhasesByAssessmentSchema(ctx, assessmentSchemaID)
	if err != nil {
		log.WithError(err).Error("Failed to get course phases by assessment schema")
		return nil, err
	}

	return coursePhaseIDs, nil
}

func CheckSchemaUsageInOtherPhases(ctx context.Context, schemaID uuid.UUID, currentCoursePhaseID uuid.UUID) (bool, error) {
	used, err := AssessmentSchemaServiceSingleton.queries.CheckAssessmentSchemaUsageInOtherPhases(ctx, db.CheckAssessmentSchemaUsageInOtherPhasesParams{
		AssessmentSchemaID: schemaID,
		CoursePhaseID:      currentCoursePhaseID,
	})
	if err != nil {
		log.WithError(err).Error("Failed to check schema usage in other phases")
		return false, err
	}

	return used, nil
}

func CopyAssessmentSchema(
	ctx context.Context,
	coursePhaseID uuid.UUID,
	sourceSchemaID uuid.UUID,
	courseIdentifierPrefix string,
) (assessmentSchemaDTO.AssessmentSchema, error) {
	tx, err := AssessmentSchemaServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return assessmentSchemaDTO.AssessmentSchema{}, err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := AssessmentSchemaServiceSingleton.queries.WithTx(tx)

	copiedSchema, err := qtx.CopyAssessmentSchema(ctx, db.CopyAssessmentSchemaParams{
		CoursePhaseID:          pgtype.UUID{Bytes: coursePhaseID, Valid: true},
		SourceSchemaID:         sourceSchemaID,
		CourseIdentifierPrefix: pgtype.Text{String: courseIdentifierPrefix, Valid: true},
	})
	if err != nil {
		log.WithError(err).Error("Failed to copy assessment schema")
		return assessmentSchemaDTO.AssessmentSchema{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return assessmentSchemaDTO.AssessmentSchema{}, err
	}

	schema := assessmentSchemaDTO.MapDBAssessmentSchemaToDTOAssessmentSchema(db.AssessmentSchema(copiedSchema))

	return schema, nil
}

func CheckSchemaOwnership(ctx context.Context, schemaID uuid.UUID, coursePhaseID uuid.UUID) (bool, error) {
	isOwner, err := AssessmentSchemaServiceSingleton.queries.CheckSchemaOwnership(ctx, db.CheckSchemaOwnershipParams{
		ID:            schemaID,
		SourcePhaseID: pgtype.UUID{Bytes: coursePhaseID, Valid: true},
	})
	if err != nil {
		log.WithError(err).Error("Failed to check schema ownership")
		return false, err
	}

	return isOwner, nil
}

func GetConsumerPhases(ctx context.Context, schemaID uuid.UUID, ownerPhaseID uuid.UUID) ([]uuid.UUID, error) {
	phases, err := AssessmentSchemaServiceSingleton.queries.GetConsumerPhases(ctx, db.GetConsumerPhasesParams{
		AssessmentSchemaID: schemaID,
		CoursePhaseID:      ownerPhaseID,
	})
	if err != nil {
		log.WithError(err).Error("Failed to get consumer phases")
		return nil, err
	}

	return phases, nil
}

func CheckPhaseHasAssessmentData(ctx context.Context, phaseID uuid.UUID, schemaID uuid.UUID) (bool, error) {
	hasData, err := AssessmentSchemaServiceSingleton.queries.CheckPhaseHasAssessmentData(ctx, db.CheckPhaseHasAssessmentDataParams{
		CoursePhaseID:      phaseID,
		AssessmentSchemaID: schemaID,
	})
	if err != nil {
		log.WithError(err).Error("Failed to check if phase has assessment data")
		return false, err
	}

	return hasData.Bool, nil
}

func UpdateAssessmentAndEvaluationCompetencies(ctx context.Context, coursePhaseID uuid.UUID, oldCompetencyID uuid.UUID, newCompetencyID uuid.UUID) error {
	tx, err := AssessmentSchemaServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := AssessmentSchemaServiceSingleton.queries.WithTx(tx)

	err = qtx.UpdateAssessmentCompetencies(ctx, db.UpdateAssessmentCompetenciesParams{
		CoursePhaseID:  coursePhaseID,
		CompetencyID:   oldCompetencyID,
		CompetencyID_2: newCompetencyID,
	})
	if err != nil {
		log.WithError(err).Error("Failed to update assessment competencies")
		return err
	}

	err = qtx.UpdateEvaluationCompetencies(ctx, db.UpdateEvaluationCompetenciesParams{
		CoursePhaseID:  coursePhaseID,
		CompetencyID:   oldCompetencyID,
		CompetencyID_2: newCompetencyID,
	})
	if err != nil {
		log.WithError(err).Error("Failed to update evaluation competencies")
		return err
	}

	return tx.Commit(ctx)
}
