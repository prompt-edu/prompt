package providerconfig

import (
	"context"
	"encoding/json"
	"fmt"

	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/encryption"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider/gitlab"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider/keycloak"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider/outline"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider/rancher"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider/slack"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/providerconfig/providerconfigDTO"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Service handles provider configuration business logic.
type Service struct {
	queries *db.Queries
}

// NewService creates a Service.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{queries: db.New(pool)}
}

// descriptorFor returns a credential-free provider used to read static metadata.
func descriptorFor(providerType string) (provider.Provider, error) {
	switch providerType {
	case "gitlab":
		return &gitlab.Provider{}, nil
	case "slack":
		return &slack.Provider{}, nil
	case "outline":
		return &outline.Provider{}, nil
	case "rancher":
		return &rancher.Provider{}, nil
	case "keycloak":
		return &keycloak.Provider{}, nil
	}
	return nil, fmt.Errorf("unknown provider type: %s", providerType)
}

// GetAuthFields returns the auth fields for a provider type.
func GetAuthFields(providerType string) ([]provider.AuthField, error) {
	descriptor, err := descriptorFor(providerType)
	if err != nil {
		return nil, err
	}
	return descriptor.GetAuthFields(), nil
}

// SupportedResourceTypes returns the resource kinds a provider type can create.
func SupportedResourceTypes(providerType string) ([]string, error) {
	descriptor, err := descriptorFor(providerType)
	if err != nil {
		return nil, err
	}
	return descriptor.SupportedResourceTypes(), nil
}

// TemplatedExtraConfigKeys returns the extra-config keys a provider treats as name
// templates, so they can be validated before they are stored.
func TemplatedExtraConfigKeys(providerType string) ([]string, error) {
	descriptor, err := descriptorFor(providerType)
	if err != nil {
		return nil, err
	}
	return descriptor.TemplatedExtraConfigKeys(), nil
}

// UpsertProviderConfig encrypts and stores the provider credentials.
func (s *Service) UpsertProviderConfig(ctx context.Context, coursePhaseID uuid.UUID, req providerconfigDTO.UpsertRequest) (providerconfigDTO.ProviderConfigResponse, error) {
	if err := validateUpsertRequest(req); err != nil {
		return providerconfigDTO.ProviderConfigResponse{}, err
	}

	raw, err := json.Marshal(req.Credentials)
	if err != nil {
		return providerconfigDTO.ProviderConfigResponse{}, fmt.Errorf("serialising credentials: %w", err)
	}

	encrypted, err := encryption.Encrypt(raw)
	if err != nil {
		return providerconfigDTO.ProviderConfigResponse{}, fmt.Errorf("encrypting credentials: %w", err)
	}

	pc, err := s.queries.UpsertProviderConfig(ctx, db.UpsertProviderConfigParams{
		CoursePhaseID: coursePhaseID,
		ProviderType:  db.ProviderType(req.ProviderType),
		Credentials:   encrypted,
	})
	if err != nil {
		return providerconfigDTO.ProviderConfigResponse{}, err
	}

	return providerconfigDTO.GetProviderConfigDTOFromDBModel(pc), nil
}

// ListProviderConfigs returns all provider configs for a phase (credentials redacted).
func (s *Service) ListProviderConfigs(ctx context.Context, coursePhaseID uuid.UUID) ([]providerconfigDTO.ProviderConfigResponse, error) {
	configs, err := s.queries.ListProviderConfigs(ctx, coursePhaseID)
	if err != nil {
		return nil, err
	}

	result := make([]providerconfigDTO.ProviderConfigResponse, len(configs))
	for i, c := range configs {
		result[i] = providerconfigDTO.GetProviderConfigDTOFromDBModel(c)
	}
	return result, nil
}

// DeleteProviderConfig removes the credentials for a provider type on a phase.
// This cascades through fk_resource_config_provider, removing all resource_config
// rows for this provider on the phase (and any resource_instance rows beneath them).
func (s *Service) DeleteProviderConfig(ctx context.Context, coursePhaseID uuid.UUID, providerType string) error {
	if err := validateProviderType(providerType); err != nil {
		return err
	}
	return s.queries.DeleteProviderConfig(ctx, db.DeleteProviderConfigParams{
		CoursePhaseID: coursePhaseID,
		ProviderType:  db.ProviderType(providerType),
	})
}

// ValidateProviderConfig decrypts credentials and calls ValidateCredentials on the provider.
func (s *Service) ValidateProviderConfig(ctx context.Context, coursePhaseID uuid.UUID, providerType string) error {
	pc, err := s.queries.GetProviderConfig(ctx, db.GetProviderConfigParams{
		CoursePhaseID: coursePhaseID,
		ProviderType:  db.ProviderType(providerType),
	})
	if err != nil {
		return fmt.Errorf("provider config not found: %w", err)
	}

	prov, err := buildProvider(db.ProviderType(providerType), pc.Credentials)
	if err != nil {
		return err
	}
	return prov.ValidateCredentials(ctx)
}

// buildProvider decrypts credentials and constructs the appropriate provider.
func buildProvider(providerType db.ProviderType, encryptedCreds []byte) (provider.Provider, error) {
	raw, err := encryption.Decrypt(encryptedCreds)
	if err != nil {
		return nil, fmt.Errorf("decrypting credentials: %w", err)
	}

	switch providerType {
	case db.ProviderTypeGitlab:
		var cfg gitlab.Config
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return nil, err
		}
		return gitlab.New(cfg), nil

	case db.ProviderTypeSlack:
		var cfg slack.Config
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return nil, err
		}
		return slack.New(cfg), nil

	case db.ProviderTypeOutline:
		var cfg outline.Config
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return nil, err
		}
		return outline.New(cfg), nil

	case db.ProviderTypeRancher:
		var cfg rancher.Config
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return nil, err
		}
		return rancher.New(cfg), nil

	case db.ProviderTypeKeycloak:
		var cfg keycloak.Config
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return nil, err
		}
		return keycloak.New(cfg), nil
	}

	return nil, fmt.Errorf("unknown provider type: %s", providerType)
}

// BuildProviderFromEncryptedCreds is exported for use by the execution worker.
func BuildProviderFromEncryptedCreds(providerType string, encryptedCreds []byte) (provider.Provider, error) {
	return buildProvider(db.ProviderType(providerType), encryptedCreds)
}
