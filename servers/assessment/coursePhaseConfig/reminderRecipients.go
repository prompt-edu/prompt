package coursePhaseConfig

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig/coursePhaseConfigDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

var getParticipationsForCoursePhaseFn = GetParticipationsForCoursePhase
var getTeamsForCoursePhaseFn = GetTeamsForCoursePhase

func GetEvaluationReminderRecipients(
	ctx context.Context,
	authHeader string,
	coursePhaseID uuid.UUID,
	evaluationType assessmentType.AssessmentType,
) (coursePhaseConfigDTO.EvaluationReminderRecipients, error) {
	config, err := CoursePhaseConfigSingleton.queries.GetCoursePhaseConfig(ctx, coursePhaseID)
	if err != nil {
		return coursePhaseConfigDTO.EvaluationReminderRecipients{}, fmt.Errorf("failed to load course phase config: %w", err)
	}

	evaluationEnabled, deadline := getEvaluationDeadlineConfig(config, evaluationType)
	deadlinePassed := !deadline.Valid || !time.Now().Before(deadline.Time)
	var deadlineTime *time.Time
	if deadline.Valid {
		deadlineCopy := deadline.Time
		deadlineTime = &deadlineCopy
	}

	participations, err := getParticipationsForCoursePhaseFn(ctx, authHeader, coursePhaseID)
	if err != nil {
		return coursePhaseConfigDTO.EvaluationReminderRecipients{}, err
	}

	result := coursePhaseConfigDTO.EvaluationReminderRecipients{
		EvaluationType:                         evaluationType,
		EvaluationTypeLabel:                    getEvaluationTypeLabel(evaluationType),
		EvaluationEnabled:                      evaluationEnabled,
		Deadline:                               deadlineTime,
		EvaluationDeadlinePlaceholder:          getEvaluationDeadlinePlaceholder(deadlineTime),
		DeadlinePassed:                         deadlinePassed,
		IncompleteAuthorCourseParticipationIDs: make([]uuid.UUID, 0),
		TotalAuthors:                           len(participations),
		CompletedAuthors:                       0,
	}

	// If the evaluation is disabled, all authors are effectively complete for reminder purposes.
	if !evaluationEnabled {
		result.CompletedAuthors = len(participations)
		return result, nil
	}

	teams, err := getTeamsForCoursePhaseFn(ctx, authHeader, coursePhaseID)
	if err != nil {
		return coursePhaseConfigDTO.EvaluationReminderRecipients{}, err
	}

	completions, err := CoursePhaseConfigSingleton.queries.GetEvaluationCompletionsByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		return coursePhaseConfigDTO.EvaluationReminderRecipients{}, fmt.Errorf("failed to load evaluation completions: %w", err)
	}

	completedTargetsByAuthor := getCompletedTargetsByAuthor(completions, evaluationType)
	teamsByID := getTeamsByID(teams)

	for _, participation := range participations {
		authorID := participation.CourseParticipationID
		if evaluationType != assessmentType.Self {
			if participation.TeamID == nil {
				result.IncompleteAuthorCourseParticipationIDs = append(
					result.IncompleteAuthorCourseParticipationIDs,
					authorID,
				)
				continue
			}
			if _, exists := teamsByID[*participation.TeamID]; !exists {
				result.IncompleteAuthorCourseParticipationIDs = append(
					result.IncompleteAuthorCourseParticipationIDs,
					authorID,
				)
				continue
			}
		}

		expectedTargets := getExpectedTargets(evaluationType, participation, teamsByID)
		if len(expectedTargets) == 0 {
			result.CompletedAuthors++
			continue
		}

		authorCompletedTargets := completedTargetsByAuthor[authorID]
		allTargetsCompleted := true
		for _, targetID := range expectedTargets {
			if _, ok := authorCompletedTargets[targetID]; !ok {
				allTargetsCompleted = false
				break
			}
		}

		if allTargetsCompleted {
			result.CompletedAuthors++
			continue
		}

		result.IncompleteAuthorCourseParticipationIDs = append(
			result.IncompleteAuthorCourseParticipationIDs,
			authorID,
		)
	}

	sort.Slice(result.IncompleteAuthorCourseParticipationIDs, func(i, j int) bool {
		return result.IncompleteAuthorCourseParticipationIDs[i].String() <
			result.IncompleteAuthorCourseParticipationIDs[j].String()
	})

	return result, nil
}

func getEvaluationDeadlineConfig(
	config db.CoursePhaseConfig,
	evaluationType assessmentType.AssessmentType,
) (bool, pgtype.Timestamptz) {
	switch evaluationType {
	case assessmentType.Self:
		return config.SelfEvaluationEnabled, config.SelfEvaluationDeadline
	case assessmentType.Peer:
		return config.PeerEvaluationEnabled, config.PeerEvaluationDeadline
	case assessmentType.Tutor:
		return config.TutorEvaluationEnabled, config.TutorEvaluationDeadline
	default:
		return false, pgtype.Timestamptz{}
	}
}

func getCompletedTargetsByAuthor(
	completions []db.EvaluationCompletion,
	evaluationType assessmentType.AssessmentType,
) map[uuid.UUID]map[uuid.UUID]struct{} {
	result := make(map[uuid.UUID]map[uuid.UUID]struct{})
	dbType := assessmentType.MapDTOtoDBAssessmentType(evaluationType)

	for _, completion := range completions {
		if completion.Type != dbType || !completion.Completed {
			continue
		}

		if _, exists := result[completion.AuthorCourseParticipationID]; !exists {
			result[completion.AuthorCourseParticipationID] = make(map[uuid.UUID]struct{})
		}
		result[completion.AuthorCourseParticipationID][completion.CourseParticipationID] = struct{}{}
	}

	return result
}

func getTeamsByID(teams []promptTypes.Team) map[uuid.UUID]promptTypes.Team {
	teamsByID := make(map[uuid.UUID]promptTypes.Team, len(teams))
	for _, team := range teams {
		teamsByID[team.ID] = team
	}
	return teamsByID
}

func getExpectedTargets(
	evaluationType assessmentType.AssessmentType,
	author coursePhaseConfigDTO.AssessmentParticipationWithStudent,
	teamsByID map[uuid.UUID]promptTypes.Team,
) []uuid.UUID {
	switch evaluationType {
	case assessmentType.Self:
		return []uuid.UUID{author.CourseParticipationID}
	case assessmentType.Peer:
		return getPeerTargets(author, teamsByID)
	case assessmentType.Tutor:
		return getTutorTargets(author, teamsByID)
	default:
		return nil
	}
}

func getPeerTargets(
	author coursePhaseConfigDTO.AssessmentParticipationWithStudent,
	teamsByID map[uuid.UUID]promptTypes.Team,
) []uuid.UUID {
	team := getTeamForAuthor(author, teamsByID)
	if team == nil {
		return nil
	}

	targets := make([]uuid.UUID, 0, len(team.Members))
	for _, member := range team.Members {
		if member.ID == author.CourseParticipationID {
			continue
		}
		targets = append(targets, member.ID)
	}

	return deduplicateUUIDs(targets)
}

func getTutorTargets(
	author coursePhaseConfigDTO.AssessmentParticipationWithStudent,
	teamsByID map[uuid.UUID]promptTypes.Team,
) []uuid.UUID {
	team := getTeamForAuthor(author, teamsByID)
	if team == nil {
		return nil
	}

	targets := make([]uuid.UUID, 0, len(team.Tutors))
	for _, tutor := range team.Tutors {
		targets = append(targets, tutor.ID)
	}

	return deduplicateUUIDs(targets)
}

func getTeamForAuthor(
	author coursePhaseConfigDTO.AssessmentParticipationWithStudent,
	teamsByID map[uuid.UUID]promptTypes.Team,
) *promptTypes.Team {
	if author.TeamID == nil {
		return nil
	}

	team, ok := teamsByID[*author.TeamID]
	if !ok {
		return nil
	}

	return &team
}

func deduplicateUUIDs(ids []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(ids))
	result := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func getEvaluationTypeLabel(evaluationType assessmentType.AssessmentType) string {
	switch evaluationType {
	case assessmentType.Self:
		return "Self Evaluation"
	case assessmentType.Peer:
		return "Peer Evaluation"
	case assessmentType.Tutor:
		return "Tutor Evaluation"
	default:
		return string(evaluationType)
	}
}

func getEvaluationDeadlinePlaceholder(deadline *time.Time) string {
	if deadline == nil {
		return ""
	}
	return deadline.Format("02.01.2006 15:04")
}
