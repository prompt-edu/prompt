package coursePhaseConfig

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig/coursePhaseConfigDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	log "github.com/sirupsen/logrus"
)

var ErrNotStarted = errors.New("assessment has not started yet")
var ErrDeadlinePassed = errors.New("deadline has passed")
var ErrCannotChangeSchemaWithData = errors.New("cannot change assessment schema when assessment or evaluation data exists")

type CoursePhaseConfigService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var CoursePhaseConfigSingleton *CoursePhaseConfigService

func NewCoursePhaseConfigService(queries db.Queries, conn *pgxpool.Pool) *CoursePhaseConfigService {
	return &CoursePhaseConfigService{
		queries: queries,
		conn:    conn,
	}
}

func validateSchemaChange(ctx context.Context, coursePhaseID, oldSchemaID, newSchemaID uuid.UUID, schemaType string) error {
	if oldSchemaID == newSchemaID {
		return nil // No change, no validation needed
	}

	hasData, err := assessmentSchemas.CheckPhaseHasAssessmentData(ctx, coursePhaseID, oldSchemaID)
	if err != nil {
		return err
	}

	if hasData {
		log.WithFields(log.Fields{
			"coursePhaseID": coursePhaseID,
			"oldSchemaID":   oldSchemaID,
			"newSchemaID":   newSchemaID,
			"schemaType":    schemaType,
		}).Errorf("Cannot change %s schema - data exists", schemaType)
		return ErrCannotChangeSchemaWithData
	}

	return nil
}

func validateSchemaAccessibility(ctx context.Context, coursePhaseID, schemaID uuid.UUID) error {
	if schemaID == uuid.Nil {
		return nil
	}

	isAccessible, err := assessmentSchemas.CheckSchemaAccessibleForCoursePhase(ctx, coursePhaseID, schemaID)
	if err != nil {
		return err
	}
	if !isAccessible {
		return assessmentSchemas.ErrSchemaNotAccessible
	}

	return nil
}

func GetCoursePhaseConfig(ctx context.Context, coursePhaseID uuid.UUID) (coursePhaseConfigDTO.CoursePhaseConfig, error) {
	config, err := CoursePhaseConfigSingleton.queries.GetCoursePhaseConfig(ctx, coursePhaseID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		tx, err := CoursePhaseConfigSingleton.conn.Begin(ctx)
		if err != nil {
			return coursePhaseConfigDTO.CoursePhaseConfig{}, err
		}
		defer promptSDK.DeferDBRollback(tx, ctx)

		qtx := CoursePhaseConfigSingleton.queries.WithTx(tx)

		err = qtx.CreateDefaultCoursePhaseConfig(ctx, coursePhaseID)
		if err != nil {
			log.WithError(err).Error("Failed to create or update course phase config")
			return coursePhaseConfigDTO.CoursePhaseConfig{}, err
		}

		err = tx.Commit(ctx)
		if err != nil {
			log.WithError(err).Error("Failed to commit transaction for course phase config creation")
			return coursePhaseConfigDTO.CoursePhaseConfig{}, err
		}
	} else if err != nil {
		log.Error("could not get course phase config: ", err)
		return coursePhaseConfigDTO.CoursePhaseConfig{}, errors.New("could not get course phase config")
	}

	return coursePhaseConfigDTO.MapDBCoursePhaseConfigToDTOCoursePhaseConfig(config), nil
}

func CreateOrUpdateCoursePhaseConfig(ctx context.Context, coursePhaseID uuid.UUID, req coursePhaseConfigDTO.CreateOrUpdateCoursePhaseConfigRequest) error {
	existingConfig, err := CoursePhaseConfigSingleton.queries.GetCoursePhaseConfig(ctx, coursePhaseID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.WithError(err).Error("Failed to get existing course phase config")
		return err
	}

	// If config exists, validate schema changes
	if err == nil {
		schemasToValidate := []struct {
			old, new   uuid.UUID
			schemaType string
		}{
			{existingConfig.AssessmentSchemaID, req.AssessmentSchemaID, "assessment"},
			{existingConfig.SelfEvaluationSchema, req.SelfEvaluationSchema, "self evaluation"},
			{existingConfig.PeerEvaluationSchema, req.PeerEvaluationSchema, "peer evaluation"},
			{existingConfig.TutorEvaluationSchema, req.TutorEvaluationSchema, "tutor evaluation"},
		}

		for _, schema := range schemasToValidate {
			if err := validateSchemaChange(ctx, coursePhaseID, schema.old, schema.new, schema.schemaType); err != nil {
				return err
			}
		}
	}

	schemaIDsToValidate := []uuid.UUID{
		req.AssessmentSchemaID,
		req.SelfEvaluationSchema,
		req.PeerEvaluationSchema,
		req.TutorEvaluationSchema,
	}
	for _, schemaID := range schemaIDsToValidate {
		if err := validateSchemaAccessibility(ctx, coursePhaseID, schemaID); err != nil {
			return err
		}
	}

	tx, err := CoursePhaseConfigSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := CoursePhaseConfigSingleton.queries.WithTx(tx)

	gradeSuggestionVisible := pgtype.Bool{}
	if req.GradeSuggestionVisible != nil {
		gradeSuggestionVisible = pgtype.Bool{Bool: *req.GradeSuggestionVisible, Valid: true}
	}

	actionItemsVisible := pgtype.Bool{}
	if req.ActionItemsVisible != nil {
		actionItemsVisible = pgtype.Bool{Bool: *req.ActionItemsVisible, Valid: true}
	}

	gradingSheetVisible := pgtype.Bool{}
	if req.GradingSheetVisible != nil {
		gradingSheetVisible = pgtype.Bool{Bool: *req.GradingSheetVisible, Valid: true}
	}

	// Preserve existing ResultsReleased value on update, default to false for new creates
	resultsReleased := pgtype.Bool{Bool: false, Valid: true}
	if err == nil {
		// Config exists - preserve the existing ResultsReleased value
		resultsReleased = pgtype.Bool{Bool: existingConfig.ResultsReleased, Valid: true}
	}

	params := db.CreateOrUpdateCoursePhaseConfigParams{
		AssessmentSchemaID:       req.AssessmentSchemaID,
		CoursePhaseID:            coursePhaseID,
		Start:                    pgtype.Timestamptz{Time: req.Start, Valid: !req.Start.IsZero()},
		Deadline:                 pgtype.Timestamptz{Time: req.Deadline, Valid: !req.Deadline.IsZero()},
		SelfEvaluationEnabled:    req.SelfEvaluationEnabled,
		SelfEvaluationSchema:     req.SelfEvaluationSchema,
		SelfEvaluationStart:      pgtype.Timestamptz{Time: req.SelfEvaluationStart, Valid: !req.SelfEvaluationStart.IsZero()},
		SelfEvaluationDeadline:   pgtype.Timestamptz{Time: req.SelfEvaluationDeadline, Valid: !req.SelfEvaluationDeadline.IsZero()},
		PeerEvaluationEnabled:    req.PeerEvaluationEnabled,
		PeerEvaluationSchema:     req.PeerEvaluationSchema,
		PeerEvaluationStart:      pgtype.Timestamptz{Time: req.PeerEvaluationStart, Valid: !req.PeerEvaluationStart.IsZero()},
		PeerEvaluationDeadline:   pgtype.Timestamptz{Time: req.PeerEvaluationDeadline, Valid: !req.PeerEvaluationDeadline.IsZero()},
		TutorEvaluationEnabled:   req.TutorEvaluationEnabled,
		TutorEvaluationSchema:    req.TutorEvaluationSchema,
		TutorEvaluationStart:     pgtype.Timestamptz{Time: req.TutorEvaluationStart, Valid: !req.TutorEvaluationStart.IsZero()},
		TutorEvaluationDeadline:  pgtype.Timestamptz{Time: req.TutorEvaluationDeadline, Valid: !req.TutorEvaluationDeadline.IsZero()},
		EvaluationResultsVisible: req.EvaluationResultsVisible,
		GradeSuggestionVisible:   gradeSuggestionVisible,
		ActionItemsVisible:       actionItemsVisible,
		ResultsReleased:          resultsReleased,
		GradingSheetVisible:      gradingSheetVisible,
	}

	err = qtx.CreateOrUpdateCoursePhaseConfig(ctx, params)
	if err != nil {
		log.WithError(err).Error("Failed to create or update course phase config")
		return err
	}

	return tx.Commit(ctx)
}

func IsAssessmentOpen(ctx context.Context, coursePhaseID uuid.UUID) (bool, error) {
	open, err := CoursePhaseConfigSingleton.queries.IsAssessmentOpen(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not check if assessment is open: ", err)
		return false, errors.New("could not check if assessment is open")
	}
	return open, nil
}

func IsAssessmentDeadlinePassed(ctx context.Context, coursePhaseID uuid.UUID) (bool, error) {
	deadlinePassed, err := CoursePhaseConfigSingleton.queries.IsAssessmentDeadlinePassed(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not check if assessment deadline has passed: ", err)
		return false, errors.New("could not check if assessment deadline has passed")
	}
	return deadlinePassed, nil
}

func IsSelfEvaluationOpen(ctx context.Context, coursePhaseID uuid.UUID) (bool, error) {
	open, err := CoursePhaseConfigSingleton.queries.IsSelfEvaluationOpen(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not check if self evaluation is open: ", err)
		return false, errors.New("could not check if self evaluation is open")
	}
	return open, nil
}

func IsSelfEvaluationDeadlinePassed(ctx context.Context, coursePhaseID uuid.UUID) (bool, error) {
	deadlinePassed, err := CoursePhaseConfigSingleton.queries.IsSelfEvaluationDeadlinePassed(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not check if self evaluation deadline has passed: ", err)
		return false, errors.New("could not check if self evaluation deadline has passed")
	}
	return deadlinePassed, nil
}

func IsPeerEvaluationOpen(ctx context.Context, coursePhaseID uuid.UUID) (bool, error) {
	open, err := CoursePhaseConfigSingleton.queries.IsPeerEvaluationOpen(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not check if peer evaluation is open: ", err)
		return false, errors.New("could not check if peer evaluation is open")
	}
	return open, nil
}

func IsPeerEvaluationDeadlinePassed(ctx context.Context, coursePhaseID uuid.UUID) (bool, error) {
	deadlinePassed, err := CoursePhaseConfigSingleton.queries.IsPeerEvaluationDeadlinePassed(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not check if peer evaluation deadline has passed: ", err)
		return false, errors.New("could not check if peer evaluation deadline has passed")
	}
	return deadlinePassed, nil
}

func IsTutorEvaluationOpen(ctx context.Context, coursePhaseID uuid.UUID) (bool, error) {
	open, err := CoursePhaseConfigSingleton.queries.IsTutorEvaluationOpen(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not check if tutor evaluation is open: ", err)
		return false, errors.New("could not check if tutor evaluation is open")
	}
	return open, nil
}

func IsTutorEvaluationDeadlinePassed(ctx context.Context, coursePhaseID uuid.UUID) (bool, error) {
	deadlinePassed, err := CoursePhaseConfigSingleton.queries.IsTutorEvaluationDeadlinePassed(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not check if tutor evaluation deadline has passed: ", err)
		return false, errors.New("could not check if tutor evaluation deadline has passed")
	}
	return deadlinePassed, nil
}

func ReleaseResults(ctx context.Context, coursePhaseID uuid.UUID) error {
	config, err := CoursePhaseConfigSingleton.queries.GetCoursePhaseConfig(ctx, coursePhaseID)
	if err != nil {
		log.WithError(err).Error("Failed to get course phase config")
		return errors.New("failed to get course phase config")
	}

	if config.ResultsReleased {
		log.WithField("coursePhaseID", coursePhaseID).Info("Results already released")
		return nil
	}

	tx, err := CoursePhaseConfigSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := CoursePhaseConfigSingleton.queries.WithTx(tx)

	params := db.CreateOrUpdateCoursePhaseConfigParams{
		AssessmentSchemaID:       config.AssessmentSchemaID,
		CoursePhaseID:            coursePhaseID,
		Start:                    config.Start,
		Deadline:                 config.Deadline,
		SelfEvaluationEnabled:    config.SelfEvaluationEnabled,
		SelfEvaluationSchema:     config.SelfEvaluationSchema,
		SelfEvaluationStart:      config.SelfEvaluationStart,
		SelfEvaluationDeadline:   config.SelfEvaluationDeadline,
		PeerEvaluationEnabled:    config.PeerEvaluationEnabled,
		PeerEvaluationSchema:     config.PeerEvaluationSchema,
		PeerEvaluationStart:      config.PeerEvaluationStart,
		PeerEvaluationDeadline:   config.PeerEvaluationDeadline,
		TutorEvaluationEnabled:   config.TutorEvaluationEnabled,
		TutorEvaluationSchema:    config.TutorEvaluationSchema,
		TutorEvaluationStart:     config.TutorEvaluationStart,
		TutorEvaluationDeadline:  config.TutorEvaluationDeadline,
		EvaluationResultsVisible: config.EvaluationResultsVisible,
		GradeSuggestionVisible:   pgtype.Bool{Bool: config.GradeSuggestionVisible, Valid: true},
		ActionItemsVisible:       pgtype.Bool{Bool: config.ActionItemsVisible, Valid: true},
		ResultsReleased:          pgtype.Bool{Bool: true, Valid: true},
		GradingSheetVisible:      pgtype.Bool{Bool: config.GradingSheetVisible, Valid: true},
	}

	err = qtx.CreateOrUpdateCoursePhaseConfig(ctx, params)
	if err != nil {
		log.WithError(err).Error("Failed to release results")
		return err
	}

	log.WithField("coursePhaseID", coursePhaseID).Info("Results released successfully")
	return tx.Commit(ctx)
}

func UpdateCoursePhaseConfigAssessmentSchema(ctx context.Context, coursePhaseID uuid.UUID, oldSchemaID uuid.UUID, newSchemaID uuid.UUID) error {
	hasData, err := assessmentSchemas.CheckPhaseHasAssessmentData(ctx, coursePhaseID, oldSchemaID)
	if err != nil {
		return err
	}
	if hasData {
		log.WithFields(log.Fields{
			"coursePhaseID": coursePhaseID,
			"oldSchemaID":   oldSchemaID,
			"newSchemaID":   newSchemaID,
		}).Error("Cannot change schema - assessment or evaluation data exists")
		return ErrCannotChangeSchemaWithData
	}

	tx, err := CoursePhaseConfigSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := CoursePhaseConfigSingleton.queries.WithTx(tx)

	// Get current config to determine which schema field to update
	config, err := qtx.GetCoursePhaseConfig(ctx, coursePhaseID)
	if err != nil {
		log.WithError(err).Error("Failed to get course phase config")
		return errors.New("failed to get course phase config")
	}
	updatedAny := false

	if oldSchemaID == config.AssessmentSchemaID {
		err = qtx.UpdateCoursePhaseConfigAssessmentSchema(ctx, db.UpdateCoursePhaseConfigAssessmentSchemaParams{
			CoursePhaseID:      coursePhaseID,
			AssessmentSchemaID: newSchemaID,
		})
		if err != nil {
			log.WithError(err).Error("Failed to update assessment schema reference")
			return err
		}
		updatedAny = true
	}

	if oldSchemaID == config.SelfEvaluationSchema {
		err = qtx.UpdateCoursePhaseConfigSelfEvaluationSchema(ctx, db.UpdateCoursePhaseConfigSelfEvaluationSchemaParams{
			CoursePhaseID:        coursePhaseID,
			SelfEvaluationSchema: newSchemaID,
		})
		if err != nil {
			log.WithError(err).Error("Failed to update self evaluation schema reference")
			return err
		}
		updatedAny = true
	}

	if oldSchemaID == config.PeerEvaluationSchema {
		err = qtx.UpdateCoursePhaseConfigPeerEvaluationSchema(ctx, db.UpdateCoursePhaseConfigPeerEvaluationSchemaParams{
			CoursePhaseID:        coursePhaseID,
			PeerEvaluationSchema: newSchemaID,
		})
		if err != nil {
			log.WithError(err).Error("Failed to update peer evaluation schema reference")
			return err
		}
		updatedAny = true
	}

	if oldSchemaID == config.TutorEvaluationSchema {
		err = qtx.UpdateCoursePhaseConfigTutorEvaluationSchema(ctx, db.UpdateCoursePhaseConfigTutorEvaluationSchemaParams{
			CoursePhaseID:         coursePhaseID,
			TutorEvaluationSchema: newSchemaID,
		})
		if err != nil {
			log.WithError(err).Error("Failed to update tutor evaluation schema reference")
			return err
		}
		updatedAny = true
	}

	if !updatedAny {
		log.WithFields(log.Fields{
			"oldSchemaID":   oldSchemaID,
			"coursePhaseID": coursePhaseID,
		}).Error("Old schema ID does not match any schema field in course phase config")
		return errors.New("old schema ID does not match any schema field in course phase config")
	}

	return tx.Commit(ctx)
}
