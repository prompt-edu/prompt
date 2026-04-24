package survey

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/team_allocation/survey/surveyDTO"
	log "github.com/sirupsen/logrus"
)

// ValidateStudentResponse validates the survey submission from a student.
func ValidateStudentResponse(ctx context.Context, coursePhaseID uuid.UUID, submission surveyDTO.StudentSurveyResponse) error {
	log.Info("validating student survey response")
	/* ----------------------------------------------------------------
	   1. Fetch course‑phase data
	-----------------------------------------------------------------*/
	preferenceMode, _ := getPreferenceModeAndTeamType(ctx, coursePhaseID)
	validTeams, err := GetActiveSurveyTeams(ctx, coursePhaseID)
	if err != nil {
		log.Error("failed to fetch teams: ", err)
		return fmt.Errorf("failed to fetch teams")
	}

	validSkills, err := SurveyServiceSingleton.queries.GetSkillsByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.Error("failed to fetch skills: ", err)
		return fmt.Errorf("failed to fetch skills")
	}

	/* ----------------------------------------------------------------
	   2. Build lookup maps
	-----------------------------------------------------------------*/
	validTeamIDs := make(map[uuid.UUID]bool, len(validTeams))
	for _, t := range validTeams {
		validTeamIDs[t.ID] = true
	}

	validSkillIDs := make(map[uuid.UUID]bool, len(validSkills))
	for _, s := range validSkills {
		validSkillIDs[s.ID] = true
	}

	/* ----------------------------------------------------------------
	   3. Validate skill responses
	-----------------------------------------------------------------*/
	// --- 3a. Quick cardinality check
	if len(submission.SkillResponses) != len(validSkills) {
		return fmt.Errorf(
			"skill responses count (%d) does not match number of available skills (%d)",
			len(submission.SkillResponses), len(validSkills),
		)
	}

	seenSkillIDs := make(map[uuid.UUID]bool, len(validSkills))
	for _, sr := range submission.SkillResponses {
		// id must be valid
		if !validSkillIDs[sr.SkillID] {
			return fmt.Errorf("skill %s is not valid for this course phase", sr.SkillID)
		}
		// no duplicates
		if seenSkillIDs[sr.SkillID] {
			return fmt.Errorf("duplicate rating found for skill %s", sr.SkillID)
		}
		seenSkillIDs[sr.SkillID] = true
	}

	// --- 3b. every valid skill appears exactly once
	// (The cardinality check + duplicate check above are enough, but the
	//   loop below keeps the intent very explicit.)
	for _, s := range validSkills {
		if !seenSkillIDs[s.ID] {
			return fmt.Errorf("missing rating for skill %s", s.ID)
		}
	}

	/* ----------------------------------------------------------------
	   4. Validate the active preference list
	-----------------------------------------------------------------*/
	if submission.PreferenceMode != "" && submission.PreferenceMode != preferenceMode {
		return fmt.Errorf("submitted preference mode %s does not match active preference mode %s", submission.PreferenceMode, preferenceMode)
	}

	preferences := submission.TeamPreferences
	preferenceLabel := "team"
	preferencePluralLabel := "teams"
	if preferenceMode == preferenceModeFields {
		if len(submission.TeamPreferences) > 0 {
			return fmt.Errorf("team preferences are not accepted in field preference mode")
		}
		preferences = submission.FieldPreferences
		preferenceLabel = "field"
		preferencePluralLabel = "fields"
	} else if len(submission.FieldPreferences) > 0 {
		return fmt.Errorf("field preferences are not accepted in team preference mode")
	}

	if len(preferences) != len(validTeams) {
		return fmt.Errorf(
			"%s preferences count (%d) does not match number of available %s (%d)",
			preferenceLabel, len(preferences), preferencePluralLabel, len(validTeams),
		)
	}

	rankMap := make(map[int]bool, len(validTeams))
	for _, tp := range preferences {
		if !validTeamIDs[tp.TeamID] {
			return fmt.Errorf("%s %s is not valid for this course phase", preferenceLabel, tp.TeamID)
		}
		if tp.Preference < 1 || int(tp.Preference) > len(validTeams) {
			return fmt.Errorf("invalid preference rank %d for %s %s", tp.Preference, preferenceLabel, tp.TeamID)
		}
		if rankMap[int(tp.Preference)] {
			return fmt.Errorf("duplicate preference rank %d found", tp.Preference)
		}
		rankMap[int(tp.Preference)] = true
	}
	for i := 1; i <= len(validTeams); i++ {
		if !rankMap[i] {
			return fmt.Errorf("missing preference rank %d", i)
		}
	}

	return nil
}
