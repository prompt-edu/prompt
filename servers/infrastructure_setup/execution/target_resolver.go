package execution

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
	log "github.com/sirupsen/logrus"
)

// ProvisioningTarget is the normalized input used by execution and providers.
type ProvisioningTarget struct {
	Scope                 db.ResourceScope
	TeamID                *uuid.UUID
	TeamName              string
	CourseParticipationID *uuid.UUID
	Student               *promptTypes.Student
	Members               []provider.Member
	TemplateData          TemplateData
}

// TargetResolver resolves PROMPT course data into provisioning targets.
type TargetResolver interface {
	ResolveTargets(ctx context.Context, authHeader string, coursePhaseID uuid.UUID, scope db.ResourceScope) ([]ProvisioningTarget, error)
	ResolveInstanceTarget(ctx context.Context, authHeader string, instance db.ResourceInstance) (ProvisioningTarget, error)
}

// CoreTargetResolver fetches targets through core's resolution endpoints. The
// upstream team / participation source is wired in the phase configurator, so
// this resolver only needs THIS phase's own ID - core merges the upstream data
// into the response automatically.
type CoreTargetResolver struct {
	queries *db.Queries
	coreURL string
}

// NewCoreTargetResolver creates the default target resolver.
func NewCoreTargetResolver(queries *db.Queries) *CoreTargetResolver {
	return &CoreTargetResolver{
		queries: queries,
		coreURL: sdkUtils.GetCoreUrl(),
	}
}

func (r *CoreTargetResolver) ResolveTargets(ctx context.Context, authHeader string, coursePhaseID uuid.UUID, scope db.ResourceScope) ([]ProvisioningTarget, error) {
	cfg, err := r.queries.GetCoursePhaseConfig(ctx, coursePhaseID)
	if err != nil {
		return nil, fmt.Errorf("infrastructure setup config missing: %w", err)
	}

	switch scope {
	case db.ResourceScopePerTeam:
		return r.resolveTeamTargets(ctx, authHeader, cfg, coursePhaseID)
	case db.ResourceScopePerStudent:
		return r.resolveStudentTargets(ctx, authHeader, cfg, coursePhaseID)
	default:
		return nil, fmt.Errorf("unsupported resource scope: %s", scope)
	}
}

func (r *CoreTargetResolver) ResolveInstanceTarget(ctx context.Context, authHeader string, instance db.ResourceInstance) (ProvisioningTarget, error) {
	config, err := r.queries.GetResourceConfig(ctx, db.GetResourceConfigParams{
		ID:            instance.ResourceConfigID,
		CoursePhaseID: instance.CoursePhaseID,
	})
	if err != nil {
		return ProvisioningTarget{}, fmt.Errorf("load resource config: %w", err)
	}

	targets, err := r.ResolveTargets(ctx, authHeader, instance.CoursePhaseID, config.Scope)
	if err != nil {
		return ProvisioningTarget{}, err
	}

	for _, target := range targets {
		if instance.TeamID != nil && target.TeamID != nil && *instance.TeamID == *target.TeamID {
			return target, nil
		}
		if instance.CourseParticipationID != nil && target.CourseParticipationID != nil && *instance.CourseParticipationID == *target.CourseParticipationID {
			return target, nil
		}
	}
	return ProvisioningTarget{}, errors.New("resource instance target no longer exists")
}

func (r *CoreTargetResolver) resolveStudentTargets(ctx context.Context, authHeader string, cfg db.CoursePhaseConfig, coursePhaseID uuid.UUID) ([]ProvisioningTarget, error) {
	participations, err := promptSDK.FetchAndMergeParticipationsWithResolutions(r.coreURL, authHeader, coursePhaseID)
	if err != nil {
		return nil, fmt.Errorf("fetch participations: %w", err)
	}

	targets := make([]ProvisioningTarget, 0, len(participations))
	for _, participation := range participations {
		participationID := participation.CourseParticipationID
		student := participation.Student
		targets = append(targets, ProvisioningTarget{
			Scope:                 db.ResourceScopePerStudent,
			CourseParticipationID: &participationID,
			Student:               &student,
			Members: []provider.Member{
				{Email: student.Email, Role: "student"},
			},
			TemplateData: TemplateData{
				StudentFirstName: student.FirstName,
				StudentLastName:  student.LastName,
				StudentEmail:     student.Email,
				StudentLogin:     student.UniversityLogin,
				StudentName:      student.FirstName + " " + student.LastName,
				SemesterTag:      cfg.SemesterTag,
				Semester:         cfg.SemesterTag,
			},
		})
	}
	return targets, nil
}

func (r *CoreTargetResolver) resolveTeamTargets(ctx context.Context, authHeader string, cfg db.CoursePhaseConfig, coursePhaseID uuid.UUID) ([]ProvisioningTarget, error) {
	phaseData, err := promptSDK.FetchAndMergeCoursePhaseWithResolution(r.coreURL, authHeader, coursePhaseID)
	if err != nil {
		return nil, fmt.Errorf("fetch phase data with resolution: %w", err)
	}

	teamsRaw, ok := phaseData["teams"]
	if !ok {
		return nil, errors.New("teams input not wired - configure the phase data graph to feed teams into this phase")
	}
	teams, err := parseTeams(teamsRaw)
	if err != nil {
		return nil, err
	}

	studentsByParticipationID, err := r.studentsByParticipationID(ctx, authHeader, coursePhaseID)
	if err != nil {
		return nil, err
	}

	targets := make([]ProvisioningTarget, 0, len(teams))
	for _, team := range teams {
		teamID := team.ID
		members := make([]provider.Member, 0, len(team.Members)+len(team.Tutors))
		for _, member := range team.Members {
			if student, ok := studentsByParticipationID[member.ID]; ok && student.Email != "" {
				members = append(members, provider.Member{Email: student.Email, Role: "student"})
			}
		}
		for _, tutor := range team.Tutors {
			if student, ok := studentsByParticipationID[tutor.ID]; ok && student.Email != "" {
				members = append(members, provider.Member{Email: student.Email, Role: "tutor"})
			}
		}

		targets = append(targets, ProvisioningTarget{
			Scope:        db.ResourceScopePerTeam,
			TeamID:       &teamID,
			TeamName:     team.Name,
			Members:      members,
			TemplateData: TemplateData{TeamName: team.Name, SemesterTag: cfg.SemesterTag, Semester: cfg.SemesterTag},
		})
	}
	return targets, nil
}

func (r *CoreTargetResolver) studentsByParticipationID(ctx context.Context, authHeader string, coursePhaseID uuid.UUID) (map[uuid.UUID]promptTypes.Student, error) {
	participations, err := promptSDK.FetchAndMergeParticipationsWithResolutions(r.coreURL, authHeader, coursePhaseID)
	if err != nil {
		return nil, fmt.Errorf("fetch source participations: %w", err)
	}

	students := make(map[uuid.UUID]promptTypes.Student, len(participations))
	for _, participation := range participations {
		students[participation.CourseParticipationID] = participation.Student
	}
	return students, nil
}

// parseTeams converts the resolved teams payload (returned as []interface{} by
// FetchAndMergeCoursePhaseWithResolution) into a typed slice. Mirrors the
// assessment phase's parser - keeping the logic local avoids a cross-service
// Go import, which the rest of the codebase also avoids.
func parseTeams(teamsRaw interface{}) ([]promptTypes.Team, error) {
	teams := make([]promptTypes.Team, 0)
	if teamsRaw == nil {
		return teams, nil
	}

	teamsSlice, ok := teamsRaw.([]interface{})
	if !ok {
		return nil, errors.New("teams field is not a slice")
	}

	for i, raw := range teamsSlice {
		team, ok := parseTeam(raw, i)
		if ok {
			teams = append(teams, team)
		}
	}
	return teams, nil
}

func parseTeam(raw interface{}, index int) (promptTypes.Team, bool) {
	m, ok := raw.(map[string]interface{})
	if !ok {
		log.Warnf("skipping team at index %d: not a map", index)
		return promptTypes.Team{}, false
	}

	idStr, ok := m["id"].(string)
	if !ok {
		log.Warnf("skipping team at index %d: missing or invalid id", index)
		return promptTypes.Team{}, false
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Warnf("skipping team at index %d: invalid uuid: %v", index, err)
		return promptTypes.Team{}, false
	}

	name, ok := m["name"].(string)
	if !ok {
		log.Warnf("skipping team at index %d: missing or invalid name", index)
		return promptTypes.Team{}, false
	}

	return promptTypes.Team{
		ID:      id,
		Name:    name,
		Members: parsePersons(m["members"]),
		Tutors:  parsePersons(m["tutors"]),
	}, true
}

func parsePersons(raw interface{}) []promptTypes.Person {
	persons := make([]promptTypes.Person, 0)
	if raw == nil {
		return persons
	}
	slice, ok := raw.([]interface{})
	if !ok {
		return persons
	}
	for _, entry := range slice {
		m, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		idStr, _ := m["id"].(string)
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		firstName, _ := m["firstName"].(string)
		lastName, _ := m["lastName"].(string)
		persons = append(persons, promptTypes.Person{ID: id, FirstName: firstName, LastName: lastName})
	}
	return persons
}
