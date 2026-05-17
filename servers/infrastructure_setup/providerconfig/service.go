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

// GetAuthFields returns the auth fields for a provider type.
func GetAuthFields(providerType string) ([]provider.AuthField, error) {
	switch providerType {
	case "gitlab":
		return (&gitlab.Provider{}).GetAuthFields(), nil
	case "slack":
		return (&slack.Provider{}).GetAuthFields(), nil
	case "outline":
		return (&outline.Provider{}).GetAuthFields(), nil
	case "rancher":
		return (&rancher.Provider{}).GetAuthFields(), nil
	case "keycloak":
		return (&keycloak.Provider{}).GetAuthFields(), nil
	}
	return nil, fmt.Errorf("unknown provider type: %s", providerType)
}

// UpsertProviderConfig encrypts and stores the provider credentials.
func (s *Service) UpsertProviderConfig(ctx context.Context, coursePhaseID uuid.UUID, req UpsertRequest) (ProviderConfigResponse, error) {
	raw, err := json.Marshal(req.Credentials)
	if err != nil {
		return ProviderConfigResponse{}, fmt.Errorf("serialising credentials: %w", err)
	}

	encrypted, err := encryption.Encrypt(raw)
	if err != nil {
		return ProviderConfigResponse{}, fmt.Errorf("encrypting credentials: %w", err)
	}

	pc, err := s.queries.UpsertProviderConfig(ctx, db.UpsertProviderConfigParams{
		CoursePhaseID: coursePhaseID,
		ProviderType:  db.ProviderType(req.ProviderType),
		Credentials:   encrypted,
	})
	if err != nil {
		return ProviderConfigResponse{}, err
	}

	return toResponse(pc), nil
}

// ListProviderConfigs returns all provider configs for a phase (credentials redacted).
func (s *Service) ListProviderConfigs(ctx context.Context, coursePhaseID uuid.UUID) ([]ProviderConfigResponse, error) {
	configs, err := s.queries.ListProviderConfigs(ctx, coursePhaseID)
	if err != nil {
		return nil, err
	}

	result := make([]ProviderConfigResponse, len(configs))
	for i, c := range configs {
		result[i] = toResponse(c)
	}
	return result, nil
}

// ValidateProviderConfig decrypts credentials and calls ValidateCredentials on the provider.
func (s *Service) ValidateProviderConfig(ctx context.Context, coursePhaseID uuid.UUID, providerType string) error {
	pc, err := s.queries.GetProviderConfig(ctx, coursePhaseID, db.ProviderType(providerType))
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

func toResponse(pc db.ProviderConfig) ProviderConfigResponse {
	return ProviderConfigResponse{
		ID:           pc.ID,
		ProviderType: string(pc.ProviderType),
	}
}

// BuildProviderFromEncryptedCreds is exported for use by the execution worker.
func BuildProviderFromEncryptedCreds(providerType string, encryptedCreds []byte) (provider.Provider, error) {
	return buildProvider(db.ProviderType(providerType), encryptedCreds)
}
