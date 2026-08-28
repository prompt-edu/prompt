// Package provider defines the Provider interface and shared types used by all
// infrastructure provider implementations (GitLab, Slack, Outline, Rancher, Keycloak).
package provider

import "context"

// Resource represents a successfully created external resource.
type Resource struct {
	ExternalID  string
	ExternalURL string
	// Warnings lists members that could not be granted access. The resource itself
	// exists, so the instance is recorded as partial rather than failed.
	Warnings []string
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
	// Only the keys a provider declares through TemplatedExtraConfigKeys have their
	// placeholders resolved; every other value reaches the provider verbatim.
	ExtraConfig map[string]interface{}
	// StableKey identifies this (resource config, target) pair independently of any
	// display name or name template. Providers use it to pair an external object with
	// PROMPT and to prove ownership before adopting a resource by name.
	StableKey string
	// ExistingExternalID is the ID already recorded for this instance, set when a
	// partial or failed instance is retried. When present a provider must re-attach by
	// ID rather than looking the resource up by name.
	ExistingExternalID string
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
	// SupportedResourceTypes returns the resource kinds this provider can create.
	SupportedResourceTypes() []string
	// TemplatedExtraConfigKeys returns the extra-config keys whose values are name
	// templates. Only these have their placeholders resolved and validated; every other
	// extra-config value stays literal, so a provider setting can still contain braces.
	TemplatedExtraConfigKeys() []string
	// CreateResource creates the external resource and returns its ID and URL.
	// Implementations must be idempotent: if the resource already exists, return it.
	// A member that cannot be granted access is reported through Resource.Warnings,
	// never silently skipped.
	CreateResource(ctx context.Context, input CreateResourceInput) (*Resource, error)
}

// PROMPT never deletes external resources. Providers adopt existing resources by name,
// so a delete could remove something the course does not own; instance deletion only
// drops the PROMPT row and external cleanup stays manual.
