package survey

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/team_allocation/survey/surveyDTO"
	log "github.com/sirupsen/logrus"
)

// SurveyService encapsulates survey-related operations.
type SurveyService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

// SurveyServiceSingleton provides a global instance.
var SurveyServiceSingleton *SurveyService

const (
	allocationProfileStandard        = "standard"
	allocationProfileProjectWeek1000 = "project_week_1000_plus"
	preferenceModeTeams              = "teams"
	preferenceModeFields             = "fields"
	teamTypeStandard                 = "standard"
	teamTypeFieldBucket              = "field_bucket"
)

func GetAllocationProfile(ctx context.Context, coursePhaseID uuid.UUID) string {
	profile, err := SurveyServiceSingleton.queries.GetAllocationProfile(ctx, coursePhaseID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Warn("could not get allocation profile, falling back to standard: ", err)
		}
		return allocationProfileStandard
	}
	if profile == allocationProfileProjectWeek1000 {
		return allocationProfileProjectWeek1000
	}
	return allocationProfileStandard
}

func SetAllocationProfile(ctx context.Context, coursePhaseID uuid.UUID, profile string) error {
	if profile != allocationProfileProjectWeek1000 {
		profile = allocationProfileStandard
	}
	return SurveyServiceSingleton.queries.UpsertAllocationProfile(ctx, db.UpsertAllocationProfileParams{
		CoursePhaseID: coursePhaseID,
		Profile:       profile,
	})
}

func getPreferenceModeAndTeamType(ctx context.Context, coursePhaseID uuid.UUID) (string, string) {
	if GetAllocationProfile(ctx, coursePhaseID) == allocationProfileProjectWeek1000 {
		return preferenceModeFields, teamTypeFieldBucket
	}
	return preferenceModeTeams, teamTypeStandard
}

func GetActiveSurveyTeams(ctx context.Context, coursePhaseID uuid.UUID) ([]db.Team, error) {
	if GetAllocationProfile(ctx, coursePhaseID) == allocationProfileProjectWeek1000 {
		return SurveyServiceSingleton.queries.GetFieldBucketTeamsByCoursePhase(ctx, coursePhaseID)
	}
	return SurveyServiceSingleton.queries.GetStandardTeamsByCoursePhase(ctx, coursePhaseID)
}

func getSurveyTimeframeForProfile(ctx context.Context, coursePhaseID uuid.UUID, profile string) (surveyDTO.SurveyTimeframe, error) {
	row := SurveyServiceSingleton.conn.QueryRow(ctx, `
		SELECT survey_start, survey_deadline
		FROM survey_timeframe_profile
		WHERE course_phase_id = $1 AND profile = $2
	`, coursePhaseID, profile)

	var surveyStart pgtype.Timestamp
	var surveyDeadline pgtype.Timestamp
	if err := row.Scan(&surveyStart, &surveyDeadline); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return surveyDTO.SurveyTimeframe{TimeframeSet: false}, nil
		}
		log.Error("could not get survey timeframe for profile: ", err)
		return surveyDTO.SurveyTimeframe{}, errors.New("could not get survey timeframe")
	}

	return surveyDTO.SurveyTimeframe{
		TimeframeSet:   true,
		SurveyStart:    surveyStart.Time,
		SurveyDeadline: surveyDeadline.Time,
	}, nil
}

// GetSurveyForm returns available teams and skills if the survey has started.
func GetSurveyForm(ctx context.Context, coursePhaseID uuid.UUID) (surveyDTO.SurveyForm, error) {
	// Get survey timeframe
	timeframe, err := GetSurveyTimeframe(ctx, coursePhaseID)
	if err != nil {
		return surveyDTO.SurveyForm{}, err
	}
	if !timeframe.TimeframeSet {
		return surveyDTO.SurveyForm{}, errors.New("could not get survey timeframe")
	}
	// Ensure survey has started.
	if time.Now().Before(timeframe.SurveyStart) {
		return surveyDTO.SurveyForm{}, errors.New("survey has not started yet")
	}
	// Get teams and skills
	preferenceMode, _ := getPreferenceModeAndTeamType(ctx, coursePhaseID)
	teams, err := GetActiveSurveyTeams(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not get teams: ", err)
		return surveyDTO.SurveyForm{}, errors.New("could not get teams")
	}
	skills, err := SurveyServiceSingleton.queries.GetSkillsByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not get skills: ", err)
		return surveyDTO.SurveyForm{}, errors.New("could not get skills")
	}
	return surveyDTO.GetSurveyDataDTOFromDBModels(teams, skills, timeframe.SurveyDeadline, preferenceMode), nil
}

// GetStudentSurveyResponses returns any submitted survey answers for the student.
func GetStudentSurveyResponses(ctx context.Context, courseParticipationID uuid.UUID, coursePhaseID uuid.UUID) (surveyDTO.StudentSurveyResponse, error) {
	preferenceMode, teamType := getPreferenceModeAndTeamType(ctx, coursePhaseID)
	preferenceResponses, err := SurveyServiceSingleton.queries.GetStudentPreferencesByTeamType(ctx, db.GetStudentPreferencesByTeamTypeParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
		TeamType:              teamType,
	})
	if err != nil {
		log.Error("could not get survey preferences: ", err)
		return surveyDTO.StudentSurveyResponse{}, errors.New("could not get survey preferences")
	}
	skillResponses, err := SurveyServiceSingleton.queries.GetStudentSkillResponsesByPreferenceMode(ctx, db.GetStudentSkillResponsesByPreferenceModeParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
		PreferenceMode:        preferenceMode,
	})
	if err != nil {
		log.Error("could not get skill responses: ", err)
		return surveyDTO.StudentSurveyResponse{}, errors.New("could not get skill responses")
	}
	return surveyDTO.GetStudentSurveyResponseDTOFromTypedDBModels(preferenceResponses, skillResponses, preferenceMode), nil
}

// SubmitSurveyResponses saves or overwrites the student's survey answers.
// It only accepts submissions before the survey_deadline.
func SubmitSurveyResponses(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID, submission surveyDTO.StudentSurveyResponse) error {
	// Get survey timeframe to check deadline.
	timeframe, err := GetSurveyTimeframe(ctx, coursePhaseID)
	if err != nil {
		return errors.New("could not get survey timeframe")
	}
	if !timeframe.TimeframeSet {
		return errors.New("could not get survey timeframe")
	}
	if time.Now().Before(timeframe.SurveyStart) {
		return errors.New("survey has not started yet")
	}
	if time.Now().After(timeframe.SurveyDeadline) {
		return errors.New("survey deadline has passed")
	}

	// Begin transaction.
	tx, err := SurveyServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := SurveyServiceSingleton.queries.WithTx(tx)

	preferenceMode, teamType := getPreferenceModeAndTeamType(ctx, coursePhaseID)
	preferences := submission.TeamPreferences
	if preferenceMode == preferenceModeFields {
		preferences = submission.FieldPreferences
	}

	// Delete only the active preference family. The other profile's answers remain available if the profile is switched back.
	if err := qtx.DeleteStudentPreferencesByTeamType(ctx, db.DeleteStudentPreferencesByTeamTypeParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
		TeamType:              teamType,
	}); err != nil {
		log.Error("failed to delete existing survey preferences: ", err)
		return errors.New("failed to delete existing survey preferences")
	}
	if err := qtx.DeleteStudentSkillResponsesByPreferenceMode(ctx, db.DeleteStudentSkillResponsesByPreferenceModeParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
		PreferenceMode:        preferenceMode,
	}); err != nil {
		log.Error("failed to delete existing skill responses: ", err)
		return errors.New("failed to delete existing skill responses")
	}

	// Insert new team preferences.
	for _, tp := range preferences {
		err := qtx.InsertStudentTeamPreference(ctx, db.InsertStudentTeamPreferenceParams{
			CourseParticipationID: courseParticipationID,
			TeamID:                tp.TeamID,
			Preference:            tp.Preference,
		})
		if err != nil {
			log.Error("failed to insert team preference: ", err)
			return errors.New("failed to insert team preference")
		}
	}

	// Insert new skill responses.
	for _, sr := range submission.SkillResponses {
		err := qtx.InsertStudentSkillResponse(ctx, db.InsertStudentSkillResponseParams{
			CourseParticipationID: courseParticipationID,
			SkillID:               sr.SkillID,
			SkillLevel:            sr.SkillLevel,
			PreferenceMode:        preferenceMode,
		})
		if err != nil {
			log.Error("failed to insert skill response: ", err)
			return errors.New("failed to insert skill response")
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("failed to commit transaction: ", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// SetSurveyTimeframe sets (or updates) the survey start and deadline for a course phase.
// It returns an error if surveyStart is not before surveyDeadline.
func SetSurveyTimeframe(ctx context.Context, coursePhaseID uuid.UUID, surveyStart, surveyDeadline time.Time) error {
	if !surveyStart.Before(surveyDeadline) {
		return errors.New("survey start must be before survey deadline")
	}

	var startTimestamp, deadlineTimestamp pgtype.Timestamp
	startTimestamp = pgtype.Timestamp{Time: surveyStart, Valid: true}
	deadlineTimestamp = pgtype.Timestamp{Time: surveyDeadline, Valid: true}

	profile := GetAllocationProfile(ctx, coursePhaseID)
	_, err := SurveyServiceSingleton.conn.Exec(ctx, `
		INSERT INTO survey_timeframe_profile (course_phase_id, profile, survey_start, survey_deadline)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (course_phase_id, profile)
		DO UPDATE SET survey_start = EXCLUDED.survey_start,
		              survey_deadline = EXCLUDED.survey_deadline
	`, coursePhaseID, profile, startTimestamp, deadlineTimestamp)
	if err != nil {
		log.Error("failed to set survey timeframe: ", err)
		return errors.New("failed to set survey timeframe")
	}
	return nil
}

func GetSurveyTimeframe(ctx context.Context, coursePhaseID uuid.UUID) (surveyDTO.SurveyTimeframe, error) {
	profile := GetAllocationProfile(ctx, coursePhaseID)
	timeframe, err := getSurveyTimeframeForProfile(ctx, coursePhaseID, profile)
	if err == nil && timeframe.TimeframeSet {
		return timeframe, nil
	}
	if err != nil {
		return surveyDTO.SurveyTimeframe{}, err
	}
	if profile != allocationProfileStandard {
		return getSurveyTimeframeForProfile(ctx, coursePhaseID, allocationProfileStandard)
	}
	return surveyDTO.SurveyTimeframe{TimeframeSet: false}, nil
}
