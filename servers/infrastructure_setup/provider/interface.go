// Package provider defines the Provider interface and shared types used by all
// infrastructure provider implementations (GitLab, Slack, Outline, Rancher, Keycloak).
package provider

import "context"

// Resource represents a successfully created external resource.
type Resource struct {
	ExternalID  string
	ExternalURL string
}

// Member represents a user to be granted access to a resource.
type Member struct {
	Email string
	Role  string // logical role: "student", "tutor", "instructor", etc.
}

// CreateResourceInput encapsulates all data needed to create a resource.
type CreateResourceInput struct {
	// Name is the resolved (sanitized) resource name.
	Name string
	// ResourceType is the provider-specific resource kind (e.g. "group", "channel").
	ResourceType string
	// Members lists the users to invite to the resource.
	Members []Member
	// PermissionMapping maps a logical role (e.g. "student") to a provider-specific
	// permission level (e.g. "developer" for GitLab, "member" for Slack).
	PermissionMapping map[string]string
	// ExtraConfig holds provider-specific configuration (e.g. Rancher's roleTemplateId).
	ExtraConfig map[string]interface{}
}

// AuthField describes a single credential field required by a provider.
type AuthField struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Type        string `json:"type"` // "text" or "password"
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// Provider is the interface every provider implementation must satisfy.
type Provider interface {
	// GetType returns the canonical provider type string (matches the DB enum).
	GetType() string
	// GetAuthFields returns the credential fields this provider requires.
	GetAuthFields() []AuthField
	// ValidateCredentials tests the credentials without creating any resource.
	ValidateCredentials(ctx context.Context) error
	// CreateResource creates the external resource and returns its ID and URL.
	// Implementations must be idempotent: if the resource already exists, return it.
	CreateResource(ctx context.Context, input CreateResourceInput) (*Resource, error)
	// DeleteResource attempts to remove the external resource identified by externalID.
	DeleteResource(ctx context.Context, externalID string) error
}
