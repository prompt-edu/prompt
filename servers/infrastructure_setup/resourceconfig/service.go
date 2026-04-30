package resourceconfig

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
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
func (s *Service) CreateResourceConfig(ctx context.Context, coursePhaseID uuid.UUID, req CreateRequest) (ResourceConfigResponse, error) {
	permJSON, err := json.Marshal(req.PermissionMapping)
	if err != nil {
		return ResourceConfigResponse{}, err
	}
	extraJSON, err := json.Marshal(req.ResourceExtraConfig)
	if err != nil {
		return ResourceConfigResponse{}, err
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
		return ResourceConfigResponse{}, err
	}
	return toResponse(rc), nil
}

// ListResourceConfigs returns all resource configs for a course phase.
func (s *Service) ListResourceConfigs(ctx context.Context, coursePhaseID uuid.UUID) ([]ResourceConfigResponse, error) {
	configs, err := s.queries.ListResourceConfigs(ctx, coursePhaseID)
	if err != nil {
		return nil, err
	}
	result := make([]ResourceConfigResponse, len(configs))
	for i, c := range configs {
		result[i] = toResponse(c)
	}
	return result, nil
}

// GetResourceConfig retrieves a single resource config.
func (s *Service) GetResourceConfig(ctx context.Context, coursePhaseID, id uuid.UUID) (ResourceConfigResponse, error) {
	rc, err := s.queries.GetResourceConfig(ctx, id, coursePhaseID)
	if err != nil {
		return ResourceConfigResponse{}, err
	}
	return toResponse(rc), nil
}

// UpdateResourceConfig updates an existing resource configuration.
func (s *Service) UpdateResourceConfig(ctx context.Context, coursePhaseID, id uuid.UUID, req UpdateRequest) (ResourceConfigResponse, error) {
	permJSON, err := json.Marshal(req.PermissionMapping)
	if err != nil {
		return ResourceConfigResponse{}, err
	}
	extraJSON, err := json.Marshal(req.ResourceExtraConfig)
	if err != nil {
		return ResourceConfigResponse{}, err
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
		return ResourceConfigResponse{}, err
	}
	return toResponse(rc), nil
}

// DeleteResourceConfig removes a resource configuration.
func (s *Service) DeleteResourceConfig(ctx context.Context, coursePhaseID, id uuid.UUID) error {
	return s.queries.DeleteResourceConfig(ctx, id, coursePhaseID)
}

func toResponse(rc db.ResourceConfig) ResourceConfigResponse {
	return ResourceConfigResponse{
		ID:                  rc.ID,
		CoursePhaseID:       rc.CoursePhaseID,
		ProviderType:        string(rc.ProviderType),
		ResourceType:        rc.ResourceType,
		Scope:               string(rc.Scope),
		NameTemplate:        rc.NameTemplate,
		PermissionMapping:   rc.PermissionMapping,
		ResourceExtraConfig: rc.ResourceExtraConfig,
		CreatedAt:           rc.CreatedAt,
	}
}
