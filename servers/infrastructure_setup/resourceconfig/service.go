package resourceconfig

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/resourceconfig/resourceconfigDTO"
)

// Service handles resource configuration business logic.
type Service struct {
	queries *db.Queries
}

// NewService creates a Service.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{queries: db.New(pool)}
}

// CreateResourceConfig persists a new resource configuration.
func (s *Service) CreateResourceConfig(ctx context.Context, coursePhaseID uuid.UUID, req resourceconfigDTO.CreateRequest) (resourceconfigDTO.ResourceConfigResponse, error) {
	if err := validateCreateResourceConfig(req); err != nil {
		return resourceconfigDTO.ResourceConfigResponse{}, err
	}
	if err := s.assertProviderConfigured(ctx, coursePhaseID, db.ProviderType(req.ProviderType)); err != nil {
		return resourceconfigDTO.ResourceConfigResponse{}, err
	}

	permJSON, err := json.Marshal(req.PermissionMapping)
	if err != nil {
		return resourceconfigDTO.ResourceConfigResponse{}, err
	}
	extraJSON, err := json.Marshal(req.ResourceExtraConfig)
	if err != nil {
		return resourceconfigDTO.ResourceConfigResponse{}, err
	}

	rc, err := s.queries.CreateResourceConfig(ctx, db.CreateResourceConfigParams{
		CoursePhaseID:       coursePhaseID,
		ProviderType:        db.ProviderType(req.ProviderType),
		ResourceType:        req.ResourceType,
		Scope:               db.ResourceScope(req.Scope),
		NameTemplate:        req.NameTemplate,
		PermissionMapping:   permJSON,
		ResourceExtraConfig: extraJSON,
	})
	if err != nil {
		if isUniqueViolation(err, resourceConfigIdentityConstraint) {
			return resourceconfigDTO.ResourceConfigResponse{}, fmt.Errorf(
				"%w: a resource configuration with this provider, resource type, scope and name template already exists in this phase",
				ErrValidation)
		}
		return resourceconfigDTO.ResourceConfigResponse{}, err
	}
	return resourceconfigDTO.GetResourceConfigDTOFromDBModel(rc), nil
}

// assertProviderConfigured rejects a resource config whose provider holds no
// credentials. A copied phase keeps the provider row but not the secret, and every
// resource pointing at it would fail on decryption at execution time.
func (s *Service) assertProviderConfigured(ctx context.Context, coursePhaseID uuid.UUID, providerType db.ProviderType) error {
	pc, err := s.queries.GetProviderConfig(ctx, db.GetProviderConfigParams{
		CoursePhaseID: coursePhaseID,
		ProviderType:  providerType,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w: %s", ErrProviderNotConfigured, providerType)
		}
		return err
	}
	if len(pc.Credentials) == 0 {
		return fmt.Errorf("%w: %s", ErrProviderNotConfigured, providerType)
	}
	return nil
}

// ListResourceConfigs returns all resource configs for a course phase.
func (s *Service) ListResourceConfigs(ctx context.Context, coursePhaseID uuid.UUID) ([]resourceconfigDTO.ResourceConfigResponse, error) {
	configs, err := s.queries.ListResourceConfigs(ctx, coursePhaseID)
	if err != nil {
		return nil, err
	}
	result := make([]resourceconfigDTO.ResourceConfigResponse, len(configs))
	for i, c := range configs {
		result[i] = resourceconfigDTO.GetResourceConfigDTOFromDBModel(c)
	}
	return result, nil
}

// GetResourceConfig retrieves a single resource config.
func (s *Service) GetResourceConfig(ctx context.Context, coursePhaseID, id uuid.UUID) (resourceconfigDTO.ResourceConfigResponse, error) {
	rc, err := s.queries.GetResourceConfig(ctx, db.GetResourceConfigParams{
		ID:            id,
		CoursePhaseID: coursePhaseID,
	})
	if err != nil {
		return resourceconfigDTO.ResourceConfigResponse{}, err
	}
	return resourceconfigDTO.GetResourceConfigDTOFromDBModel(rc), nil
}

// UpdateResourceConfig updates an existing resource configuration.
func (s *Service) UpdateResourceConfig(ctx context.Context, coursePhaseID, id uuid.UUID, req resourceconfigDTO.UpdateRequest) (resourceconfigDTO.ResourceConfigResponse, error) {
	existing, err := s.queries.GetResourceConfig(ctx, db.GetResourceConfigParams{
		ID:            id,
		CoursePhaseID: coursePhaseID,
	})
	if err != nil {
		return resourceconfigDTO.ResourceConfigResponse{}, err
	}
	if err := validateUpdateResourceConfig(string(existing.ProviderType), req); err != nil {
		return resourceconfigDTO.ResourceConfigResponse{}, err
	}
	if err := s.assertProviderConfigured(ctx, coursePhaseID, existing.ProviderType); err != nil {
		return resourceconfigDTO.ResourceConfigResponse{}, err
	}
	if err := s.assertNoLiveInstances(ctx, id); err != nil {
		return resourceconfigDTO.ResourceConfigResponse{}, err
	}

	permJSON, err := json.Marshal(req.PermissionMapping)
	if err != nil {
		return resourceconfigDTO.ResourceConfigResponse{}, err
	}
	extraJSON, err := json.Marshal(req.ResourceExtraConfig)
	if err != nil {
		return resourceconfigDTO.ResourceConfigResponse{}, err
	}

	rc, err := s.queries.UpdateResourceConfig(ctx, db.UpdateResourceConfigParams{
		ID:                  id,
		CoursePhaseID:       coursePhaseID,
		ResourceType:        req.ResourceType,
		Scope:               db.ResourceScope(req.Scope),
		NameTemplate:        req.NameTemplate,
		PermissionMapping:   permJSON,
		ResourceExtraConfig: extraJSON,
	})
	if err != nil {
		if isUniqueViolation(err, resourceConfigIdentityConstraint) {
			return resourceconfigDTO.ResourceConfigResponse{}, fmt.Errorf(
				"%w: another resource configuration in this phase already has this provider, resource type, scope and name template",
				ErrValidation)
		}
		return resourceconfigDTO.ResourceConfigResponse{}, err
	}
	return resourceconfigDTO.GetResourceConfigDTOFromDBModel(rc), nil
}

// assertNoLiveInstances refuses to edit a config that has already provisioned something.
//
// A non-failed instance blocks a second instance for the same target, so an edited
// config would never re-provision: the row would describe one resource while a
// differently named one existed externally, with no way to tell from the UI. Changing
// the permission mapping is refused for the same reason, since it would not reach the
// memberships already granted.
func (s *Service) assertNoLiveInstances(ctx context.Context, resourceConfigID uuid.UUID) error {
	live, err := s.queries.CountLiveInstancesForConfig(ctx, resourceConfigID)
	if err != nil {
		return err
	}
	if live > 0 {
		return fmt.Errorf("%w: this configuration has %d provisioned instance(s); delete them before editing it",
			ErrValidation, live)
	}
	return nil
}

// DeleteResourceConfig removes a resource configuration.
//
// The delete cascades to the config's instances, and those rows are the only record
// PROMPT keeps of the external resources: it never deletes them, so the GitLab groups
// and Slack channels stay behind with nothing pointing at them. A config that still has
// provisioned instances is therefore only deleted when the caller says so explicitly.
func (s *Service) DeleteResourceConfig(ctx context.Context, coursePhaseID, id uuid.UUID, confirmed bool) error {
	if !confirmed {
		live, err := s.queries.CountLiveInstancesForConfig(ctx, id)
		if err != nil {
			return err
		}
		if live > 0 {
			return fmt.Errorf("%w: this configuration has %d provisioned instance(s); deleting it drops PROMPT's record of the external resources they point at, and those are not removed",
				ErrConfirmationRequired, live)
		}
	}

	return s.queries.DeleteResourceConfig(ctx, db.DeleteResourceConfigParams{
		ID:            id,
		CoursePhaseID: coursePhaseID,
	})
}

// isUniqueViolation reports whether err is a Postgres unique violation on a named
// constraint, so a duplicate can be answered as a bad request rather than a 500.
func isUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode && pgErr.ConstraintName == constraint
}
