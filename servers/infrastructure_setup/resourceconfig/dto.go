package resourceconfig

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CreateRequest is the request body for creating a resource configuration.
type CreateRequest struct {
	ProviderType        string                 `json:"providerType"        binding:"required"`
	ResourceType        string                 `json:"resourceType"        binding:"required"`
	Scope               string                 `json:"scope"               binding:"required,oneof=per_team per_student"`
	NameTemplate        string                 `json:"nameTemplate"        binding:"required"`
	PermissionMapping   map[string]string      `json:"permissionMapping"`
	ResourceExtraConfig map[string]interface{} `json:"resourceExtraConfig"`
}

// UpdateRequest is the request body for updating a resource configuration.
type UpdateRequest struct {
	ResourceType        string                 `json:"resourceType"        binding:"required"`
	Scope               string                 `json:"scope"               binding:"required,oneof=per_team per_student"`
	NameTemplate        string                 `json:"nameTemplate"        binding:"required"`
	PermissionMapping   map[string]string      `json:"permissionMapping"`
	ResourceExtraConfig map[string]interface{} `json:"resourceExtraConfig"`
}

// ResourceConfigResponse is the API response for a resource configuration.
type ResourceConfigResponse struct {
	ID                  uuid.UUID       `json:"id"`
	CoursePhaseID       uuid.UUID       `json:"coursePhaseId"`
	ProviderType        string          `json:"providerType"`
	ResourceType        string          `json:"resourceType"`
	Scope               string          `json:"scope"`
	NameTemplate        string          `json:"nameTemplate"`
	PermissionMapping   json.RawMessage `json:"permissionMapping"`
	ResourceExtraConfig json.RawMessage `json:"resourceExtraConfig"`
	CreatedAt           time.Time       `json:"createdAt"`
}
