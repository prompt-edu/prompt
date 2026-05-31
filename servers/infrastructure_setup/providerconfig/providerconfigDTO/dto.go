package providerconfigDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
)

// AuthField describes a credential field required by an infrastructure provider.
type AuthField = provider.AuthField

// UpsertRequest is the request body for saving provider credentials.
type UpsertRequest struct {
	ProviderType string                 `json:"providerType" binding:"required"`
	Credentials  map[string]interface{} `json:"credentials"  binding:"required"`
}

// ProviderConfigResponse is the API response for a provider config (credentials redacted).
type ProviderConfigResponse struct {
	ID           uuid.UUID `json:"id"`
	ProviderType string    `json:"providerType"`
}

// GetProviderConfigDTOFromDBModel builds the redacted API response from the DB row.
func GetProviderConfigDTOFromDBModel(pc db.ProviderConfig) ProviderConfigResponse {
	return ProviderConfigResponse{
		ID:           pc.ID,
		ProviderType: string(pc.ProviderType),
	}
}
