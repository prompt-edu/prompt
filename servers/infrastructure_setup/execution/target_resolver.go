package execution

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
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

// CoreTargetResolver fetches targets through core and phase resolution endpoints.
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
		if cfg.TeamSourceCoursePhaseID == nil {
			return nil, errors.New("team source course phase is not configured")
		}
		return r.resolveTeamTargets(ctx, authHeader, cfg, *cfg.TeamSourceCoursePhaseID)
	case db.ResourceScopePerStudent:
		sourceID := cfg.StudentSourceCoursePhaseID
		if sourceID == nil {
			return nil, errors.New("student source course phase is not configured")
		}
		return r.resolveStudentTargets(ctx, authHeader, cfg, *sourceID)
	default:
		return nil, fmt.Errorf("unsupported resource scope: %s", scope)
	}
}

func (r *CoreTargetResolver) ResolveInstanceTarget(ctx context.Context, authHeader string, instance db.ResourceInstance) (ProvisioningTarget, error) {
	config, err := r.queries.GetResourceConfig(ctx, instance.ResourceConfigID, instance.CoursePhaseID)
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

func (r *CoreTargetResolver) resolveStudentTargets(ctx context.Context, authHeader string, cfg db.CoursePhaseConfig, sourceCoursePhaseID uuid.UUID) ([]ProvisioningTarget, error) {
	participations, err := promptSDK.FetchAndMergeParticipationsWithResolutions(r.coreURL, authHeader, sourceCoursePhaseID)
	if err != nil {
		return nil, fmt.Errorf("fetch student targets: %w", err)
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

func (r *CoreTargetResolver) resolveTeamTargets(ctx context.Context, authHeader string, cfg db.CoursePhaseConfig, sourceCoursePhaseID uuid.UUID) ([]ProvisioningTarget, error) {
	phaseData, err := promptSDK.FetchAndMergeCoursePhaseWithResolution(r.coreURL, authHeader, sourceCoursePhaseID)
	if err != nil {
		return nil, fmt.Errorf("fetch team source phase data: %w", err)
	}

	teamsRaw, ok := phaseData["teams"]
	if !ok {
		return nil, errors.New("team source phase did not expose teams")
	}

	teamsBytes, err := json.Marshal(teamsRaw)
	if err != nil {
		return nil, fmt.Errorf("marshal teams: %w", err)
	}

	var teams []promptTypes.Team
	if err := json.Unmarshal(teamsBytes, &teams); err != nil {
		return nil, fmt.Errorf("unmarshal teams: %w", err)
	}

	studentSourceID := sourceCoursePhaseID
	if cfg.StudentSourceCoursePhaseID != nil {
		studentSourceID = *cfg.StudentSourceCoursePhaseID
	}
	studentsByParticipationID, err := r.studentsByParticipationID(ctx, authHeader, studentSourceID)
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
