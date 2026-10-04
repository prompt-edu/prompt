package execution

import (
	"time"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
)

// ResourceInstanceResponse is the API response for a provisioned resource instance.
//
// It is spelled out rather than aliased to the sqlc row so a column added or renamed in
// a migration does not silently change the wire format the remote reads.
type ResourceInstanceResponse struct {
	ID                    uuid.UUID         `json:"id"`
	ResourceConfigID      uuid.UUID         `json:"resourceConfigId"`
	CoursePhaseID         uuid.UUID         `json:"coursePhaseId"`
	TeamID                *uuid.UUID        `json:"teamId"`
	CourseParticipationID *uuid.UUID        `json:"courseParticipationId"`
	Status                db.ResourceStatus `json:"status"`
	ExternalID            *string           `json:"externalId"`
	ExternalUrl           *string           `json:"externalUrl"`
	ErrorMessage          *string           `json:"errorMessage"`
	// TargetName names the team or student the row is about, as core resolved it when the
	// instance last ran. Empty until then.
	TargetName string `json:"targetName"`
	// ResolvedName is the name the provider was asked to create, so a failed row still
	// says which resource it was.
	ResolvedName string           `json:"resolvedName"`
	ProviderType db.ProviderType  `json:"providerType"`
	ResourceType string           `json:"resourceType"`
	Scope        db.ResourceScope `json:"scope"`
	NameTemplate string           `json:"nameTemplate"`
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
}

// GetResourceInstanceDTOFromDBModel maps one listed instance onto its API response.
func GetResourceInstanceDTOFromDBModel(instance db.ListResourceInstancesWithConfigRow) ResourceInstanceResponse {
	return ResourceInstanceResponse{
		ID:                    instance.ID,
		ResourceConfigID:      instance.ResourceConfigID,
		CoursePhaseID:         instance.CoursePhaseID,
		TeamID:                instance.TeamID,
		CourseParticipationID: instance.CourseParticipationID,
		Status:                instance.Status,
		ExternalID:            instance.ExternalID,
		ExternalUrl:           instance.ExternalUrl,
		ErrorMessage:          instance.ErrorMessage,
		TargetName:            instance.TargetName,
		ResolvedName:          instance.ResolvedName,
		ProviderType:          instance.ProviderType,
		ResourceType:          instance.ResourceType,
		Scope:                 instance.Scope,
		NameTemplate:          instance.NameTemplate,
		CreatedAt:             instance.CreatedAt,
		UpdatedAt:             instance.UpdatedAt,
	}
}

// GetResourceInstanceDTOsFromDBModels maps listed instances onto their API responses.
// The slice is never nil, so the endpoint answers with [] rather than null.
func GetResourceInstanceDTOsFromDBModels(instances []db.ListResourceInstancesWithConfigRow) []ResourceInstanceResponse {
	responses := make([]ResourceInstanceResponse, 0, len(instances))
	for _, instance := range instances {
		responses = append(responses, GetResourceInstanceDTOFromDBModel(instance))
	}
	return responses
}
