package execution

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
)

// ErrExecutionInProgress is returned when a run is already queued or running for the phase.
var ErrExecutionInProgress = errors.New("an execution is already in progress for this course phase")

// ErrInstanceNotRetryable is returned when an instance exists but is not in a retryable state.
var ErrInstanceNotRetryable = errors.New("only failed or partial instances can be retried")

// ErrInstanceNotFound is returned when the instance does not exist in the course phase.
var ErrInstanceNotFound = errors.New("resource instance not found")

// ErrProviderNotConfigured is returned when a referenced provider has no credentials.
var ErrProviderNotConfigured = errors.New("provider credentials are missing")

// ErrNothingConfigured is returned when the phase has no resource config to provision.
var ErrNothingConfigured = errors.New("no resource is configured for this course phase")

// ErrSemesterTagMissing is returned when a template needs the semester tag and the phase
// has none saved.
var ErrSemesterTagMissing = errors.New("a template uses {{semesterTag}}, but this phase has no semester tag saved; set it on the Configuration page")

// ErrTeamsNotWired is returned when a per_team config runs in a phase that no teams
// reach through the course's phase graph.
var ErrTeamsNotWired = errors.New("no teams reach this phase: connect a Team Allocation or Self Team Allocation phase to it in the course configurator")

// TriggerSummary reports what one trigger did. A run's most common first outcome is a
// mix of successes and failures, so the endpoint has to say whether pressing the button
// again actually queued anything rather than reporting success over a no-op.
type TriggerSummary struct {
	// Queued counts targets that had no instance yet.
	Queued int `json:"queued"`
	// Requeued counts failed or partial instances reset for another attempt.
	Requeued int `json:"requeued"`
	// UpToDate counts targets whose resource was already created.
	UpToDate int `json:"upToDate"`
}

// Started reports whether the trigger handed any work to the worker.
func (s TriggerSummary) Started() bool {
	return s.Queued+s.Requeued > 0
}

// instanceKey identifies the (resource config, target) pair one instance stands for.
// A config has exactly one scope, so a target is either a team or a participation.
type instanceKey struct {
	configID uuid.UUID
	targetID uuid.UUID
}

func targetKey(configID uuid.UUID, teamID, courseParticipationID *uuid.UUID) instanceKey {
	switch {
	case teamID != nil:
		return instanceKey{configID: configID, targetID: *teamID}
	case courseParticipationID != nil:
		return instanceKey{configID: configID, targetID: *courseParticipationID}
	default:
		return instanceKey{configID: configID}
	}
}

// Service handles resource instance lifecycle.
type Service struct {
	queries  *db.Queries
	pool     *pgxpool.Pool
	worker   *Worker
	resolver TargetResolver
}

// NewService creates a Service with a background worker.
func NewService(pool *pgxpool.Pool) *Service {
	queries := db.New(pool)
	return &Service{
		queries:  queries,
		pool:     pool,
		resolver: NewCoreTargetResolver(queries),
		worker:   NewWorker(pool),
	}
}

// NewServiceWithResolver creates a Service with an injected resolver for tests.
func NewServiceWithResolver(pool *pgxpool.Pool, resolver TargetResolver) *Service {
	queries := db.New(pool)
	return &Service{
		queries:  queries,
		pool:     pool,
		resolver: resolver,
		worker:   NewWorkerWithResolver(pool, resolver),
	}
}

// TriggerExecution converges the phase's resource instances on its resource configs and
// then starts the async worker. A target without an instance gets one, a failed or
// partial instance is queued for another attempt, and an already created resource is
// left alone.
//
// Targets are resolved first, outside the transaction, because resolution calls core
// over HTTP. The transaction then takes a per-phase advisory lock so the in-progress
// check and the writes cannot interleave with a second trigger.
func (s *Service) TriggerExecution(ctx context.Context, authHeader string, coursePhaseID uuid.UUID) (TriggerSummary, error) {
	configs, targetsByScope, err := s.prepareRun(ctx, authHeader, coursePhaseID)
	if err != nil {
		return TriggerSummary{}, err
	}

	summary, err := s.queueInstances(ctx, coursePhaseID, configs, targetsByScope)
	if err != nil {
		return summary, err
	}

	if summary.Started() {
		s.worker.RunPendingInstances(authHeader, coursePhaseID)
	}
	return summary, nil
}

// PreviewExecution reports what TriggerExecution would do right now, refusing for the
// same reasons, without writing anything. It is what lets the provisioning page say
// how many resources a click creates before anyone clicks.
func (s *Service) PreviewExecution(ctx context.Context, authHeader string, coursePhaseID uuid.UUID) (ProvisioningPreview, error) {
	var preview ProvisioningPreview

	configs, targetsByScope, err := s.prepareRun(ctx, authHeader, coursePhaseID)
	if err != nil {
		return preview, err
	}

	existing, err := s.instancesByTarget(ctx, s.queries, coursePhaseID)
	if err != nil {
		return preview, err
	}
	running, err := s.queries.CountNonTerminalInstances(ctx, coursePhaseID)
	if err != nil {
		return preview, err
	}

	plan := planInstances(coursePhaseID, configs, targetsByScope, existing)
	preview.Queued = len(plan.create)
	preview.Requeued = len(plan.requeue)
	preview.UpToDate = plan.upToDate
	preview.Running = int(running)
	if targets, ok := targetsByScope[db.ResourceScopePerTeam]; ok {
		count := len(targets)
		preview.Teams = &count
	}
	if targets, ok := targetsByScope[db.ResourceScopePerStudent]; ok {
		count := len(targets)
		preview.Students = &count
	}
	return preview, nil
}

// prepareRun loads the phase's resource configs, refuses a run that could not succeed,
// and resolves the targets of every scope in use.
func (s *Service) prepareRun(ctx context.Context, authHeader string, coursePhaseID uuid.UUID) ([]db.ResourceConfig, map[db.ResourceScope][]ProvisioningTarget, error) {
	configs, err := s.queries.ListResourceConfigs(ctx, coursePhaseID)
	if err != nil {
		return nil, nil, err
	}
	if len(configs) == 0 {
		return nil, nil, ErrNothingConfigured
	}

	if err := s.assertProvidersConfigured(ctx, coursePhaseID, configs); err != nil {
		return nil, nil, err
	}

	if err := s.assertSemesterTagAvailable(ctx, coursePhaseID, configs); err != nil {
		return nil, nil, err
	}

	targetsByScope, err := s.resolveScopes(ctx, authHeader, coursePhaseID, configs)
	if err != nil {
		return nil, nil, err
	}
	return configs, targetsByScope, nil
}

// assertProvidersConfigured rejects a run whose providers lost their credentials, which
// is the state a copied phase starts in. Without this the instances would all be created
// and then fail one by one on decryption.
func (s *Service) assertProvidersConfigured(ctx context.Context, coursePhaseID uuid.UUID, configs []db.ResourceConfig) error {
	configured, err := s.queries.ListConfiguredProviderConfigs(ctx, coursePhaseID)
	if err != nil {
		return err
	}
	available := make(map[db.ProviderType]struct{}, len(configured))
	for _, pc := range configured {
		available[pc.ProviderType] = struct{}{}
	}

	for _, cfg := range configs {
		if _, ok := available[cfg.ProviderType]; !ok {
			return fmt.Errorf("%w: %s", ErrProviderNotConfigured, cfg.ProviderType)
		}
	}
	return nil
}

// assertSemesterTagAvailable refuses to run templates that need a semester tag the phase
// does not have.
//
// Resolution replaces a placeholder it cannot fill with an empty string, so
// "{{semesterTag}}-{{teamName}}" would name a real GitLab group "-ios-team-1" (or
// "ios-team-1", once a provider trims the leading separator). The setup page prefills the
// field from the course, which makes it easy to believe a tag is stored when nothing was
// ever saved.
func (s *Service) assertSemesterTagAvailable(ctx context.Context, coursePhaseID uuid.UUID, configs []db.ResourceConfig) error {
	if !needsSemesterTag(configs) {
		return nil
	}

	phaseConfig, err := s.queries.GetCoursePhaseConfig(ctx, coursePhaseID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrSemesterTagMissing
	}
	if err != nil {
		return err
	}
	if phaseConfig.SemesterTag == "" {
		return ErrSemesterTagMissing
	}
	return nil
}

// needsSemesterTag reports whether any config would resolve the placeholder, in its name
// template or in an extra-config value a provider treats as a template.
func needsSemesterTag(configs []db.ResourceConfig) bool {
	for _, cfg := range configs {
		if UsesPlaceholder(cfg.NameTemplate, SemesterTagPlaceholder) {
			return true
		}
		extra, err := ParseExtraConfig(cfg.ResourceExtraConfig)
		if err != nil {
			continue
		}
		for _, value := range extra {
			if text, ok := value.(string); ok && UsesPlaceholder(text, SemesterTagPlaceholder) {
				return true
			}
		}
	}
	return false
}

func (s *Service) resolveScopes(ctx context.Context, authHeader string, coursePhaseID uuid.UUID, configs []db.ResourceConfig) (map[db.ResourceScope][]ProvisioningTarget, error) {
	targetsByScope := make(map[db.ResourceScope][]ProvisioningTarget)
	for _, cfg := range configs {
		if _, done := targetsByScope[cfg.Scope]; done {
			continue
		}
		targets, err := s.resolver.ResolveTargets(ctx, authHeader, coursePhaseID, cfg.Scope)
		if err != nil {
			return nil, err
		}
		targetsByScope[cfg.Scope] = targets
	}
	return targetsByScope, nil
}

// queueInstances brings the phase's instances in line with its configs and targets.
// Every (config, target) pair carries exactly one instance for the lifetime of the
// phase, so a re-trigger converges on the row that is already there instead of adding a
// second one beside it.
func (s *Service) queueInstances(ctx context.Context, coursePhaseID uuid.UUID, configs []db.ResourceConfig, targetsByScope map[db.ResourceScope][]ProvisioningTarget) (TriggerSummary, error) {
	var summary TriggerSummary

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return summary, err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	locked, err := qtx.TryLockPhaseExecution(ctx, coursePhaseID.String())
	if err != nil {
		return summary, err
	}
	if !locked {
		return summary, ErrExecutionInProgress
	}

	nonTerminal, err := qtx.CountNonTerminalInstances(ctx, coursePhaseID)
	if err != nil {
		return summary, err
	}
	if nonTerminal > 0 {
		return summary, ErrExecutionInProgress
	}

	existing, err := s.instancesByTarget(ctx, qtx, coursePhaseID)
	if err != nil {
		return summary, err
	}

	plan := planInstances(coursePhaseID, configs, targetsByScope, existing)
	for _, params := range plan.create {
		if _, err := qtx.CreateResourceInstance(ctx, params); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return summary, err
		}
		summary.Queued++
	}
	for _, instanceID := range plan.requeue {
		if _, err := qtx.ResetInstanceToPending(ctx, db.ResetInstanceToPendingParams{
			ID:            instanceID,
			CoursePhaseID: coursePhaseID,
		}); err != nil {
			return summary, err
		}
		summary.Requeued++
	}
	summary.UpToDate = plan.upToDate

	return summary, tx.Commit(ctx)
}

// instancePlan is what converging the phase's instances on its configs and targets
// takes: instances to create, failed or partial ones to queue again, and how many
// targets are already provisioned.
type instancePlan struct {
	create   []db.CreateResourceInstanceParams
	requeue  []uuid.UUID
	upToDate int
}

// planInstances decides, for every (config, target) pair, what a trigger does. It
// writes nothing, so the preview and the trigger share one answer.
func planInstances(coursePhaseID uuid.UUID, configs []db.ResourceConfig, targetsByScope map[db.ResourceScope][]ProvisioningTarget, existing map[instanceKey]db.ResourceInstance) instancePlan {
	var plan instancePlan
	for _, cfg := range configs {
		for _, target := range targetsByScope[cfg.Scope] {
			instance, ok := existing[targetKey(cfg.ID, target.TeamID, target.CourseParticipationID)]
			switch {
			case !ok:
				plan.create = append(plan.create, createResourceInstanceParams(cfg, coursePhaseID, target))
			case isRetryable(instance.Status):
				plan.requeue = append(plan.requeue, instance.ID)
			default:
				plan.upToDate++
			}
		}
	}
	return plan
}

func (s *Service) instancesByTarget(ctx context.Context, qtx *db.Queries, coursePhaseID uuid.UUID) (map[instanceKey]db.ResourceInstance, error) {
	instances, err := qtx.ListResourceInstances(ctx, coursePhaseID)
	if err != nil {
		return nil, err
	}
	byTarget := make(map[instanceKey]db.ResourceInstance, len(instances))
	for _, instance := range instances {
		byTarget[targetKey(instance.ResourceConfigID, instance.TeamID, instance.CourseParticipationID)] = instance
	}
	return byTarget, nil
}

// isRetryable mirrors ResetInstanceToPending's WHERE clause.
func isRetryable(status db.ResourceStatus) bool {
	return status == db.ResourceStatusFailed || status == db.ResourceStatusPartial
}

func createResourceInstanceParams(cfg db.ResourceConfig, coursePhaseID uuid.UUID, target ProvisioningTarget) db.CreateResourceInstanceParams {
	return db.CreateResourceInstanceParams{
		ResourceConfigID:      cfg.ID,
		CoursePhaseID:         coursePhaseID,
		TeamID:                target.TeamID,
		CourseParticipationID: target.CourseParticipationID,
		TargetName:            target.DisplayName(),
	}
}

// StartStaleClaimSweeper recovers instances a crashed process left claimed. See
// Worker.StartStaleClaimSweeper.
func (s *Service) StartStaleClaimSweeper(ctx context.Context) {
	s.worker.StartStaleClaimSweeper(ctx)
}

// ListInstances returns all resource instances for a course phase, each with the people
// its latest run was for. The slice is never nil, so the endpoint answers with [] rather
// than null when nothing is provisioned.
func (s *Service) ListInstances(ctx context.Context, coursePhaseID uuid.UUID) ([]ResourceInstanceResponse, error) {
	instances, err := s.queries.ListResourceInstancesWithConfig(ctx, coursePhaseID)
	if err != nil {
		return nil, err
	}
	members, err := s.queries.ListInstanceMembersByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		return nil, err
	}
	return GetResourceInstanceDTOsFromDBModels(instances, members), nil
}

// ListMyResources returns every resource config of the phase as the given student sees
// it: each instance provisioned for them or their team, and one entry without a status
// for a config that has provisioned nothing for them yet. The slice is never nil.
func (s *Service) ListMyResources(ctx context.Context, coursePhaseID, courseParticipationID uuid.UUID) ([]MyResourceResponse, error) {
	configs, err := s.queries.ListResourceConfigs(ctx, coursePhaseID)
	if err != nil {
		return nil, err
	}
	instances, err := s.queries.ListInstancesForParticipant(ctx, db.ListInstancesForParticipantParams{
		CoursePhaseID:         coursePhaseID,
		CourseParticipationID: courseParticipationID,
	})
	if err != nil {
		return nil, err
	}

	instancesByConfig := make(map[uuid.UUID][]db.ListInstancesForParticipantRow, len(configs))
	for _, instance := range instances {
		instancesByConfig[instance.ResourceConfigID] = append(instancesByConfig[instance.ResourceConfigID], instance)
	}

	resources := make([]MyResourceResponse, 0, len(configs))
	for _, cfg := range configs {
		entry := MyResourceResponse{
			ResourceConfigID: cfg.ID,
			ProviderType:     cfg.ProviderType,
			ResourceType:     cfg.ResourceType,
			Scope:            cfg.Scope,
		}
		matches := instancesByConfig[cfg.ID]
		if len(matches) == 0 {
			resources = append(resources, entry)
			continue
		}
		// Someone on two teams sees both teams' resources.
		for _, instance := range matches {
			status := instance.Status
			provisioned := entry
			provisioned.Status = &status
			provisioned.Granted = instance.Granted
			provisioned.Name = instance.ResolvedName
			provisioned.URL = studentFacingURL(cfg.ProviderType, instance.ExternalUrl)
			if cfg.Scope == db.ResourceScopePerTeam {
				provisioned.TeamName = instance.TargetName
			}
			resources = append(resources, provisioned)
		}
	}
	return resources, nil
}

// studentFacingURL drops links a student cannot use. Keycloak's points into the admin
// console, where a student has no account: a Keycloak group only matters through the
// services that sign in with it.
func studentFacingURL(providerType db.ProviderType, url *string) *string {
	if providerType == db.ProviderTypeKeycloak || url == nil || *url == "" {
		return nil
	}
	return url
}

// RetryInstance resets a failed or partial instance back to pending and starts the worker.
// A worker is only started when a row was actually reset.
func (s *Service) RetryInstance(ctx context.Context, authHeader string, coursePhaseID, instanceID uuid.UUID) error {
	_, err := s.queries.ResetInstanceToPending(ctx, db.ResetInstanceToPendingParams{
		ID:            instanceID,
		CoursePhaseID: coursePhaseID,
	})
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		// No row was reset: either the instance does not exist, or it is in a state
		// that cannot be retried.
		if _, getErr := s.queries.GetResourceInstance(ctx, db.GetResourceInstanceParams{
			ID:            instanceID,
			CoursePhaseID: coursePhaseID,
		}); getErr != nil {
			if errors.Is(getErr, pgx.ErrNoRows) {
				return ErrInstanceNotFound
			}
			return getErr
		}
		return ErrInstanceNotRetryable
	}

	s.worker.RunPendingInstances(authHeader, coursePhaseID)
	return nil
}

// DeleteInstance removes a resource instance. The external resource is never touched:
// providers adopt resources by name, so PROMPT cannot know that it owns them.
func (s *Service) DeleteInstance(ctx context.Context, coursePhaseID, instanceID uuid.UUID) error {
	return s.queries.DeleteResourceInstance(ctx, db.DeleteResourceInstanceParams{
		ID:            instanceID,
		CoursePhaseID: coursePhaseID,
	})
}
