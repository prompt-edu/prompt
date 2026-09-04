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
	CreatedAt             time.Time         `json:"createdAt"`
	UpdatedAt             time.Time         `json:"updatedAt"`
}

// GetResourceInstanceDTOFromDBModel maps one stored instance onto its API response.
func GetResourceInstanceDTOFromDBModel(instance db.ResourceInstance) ResourceInstanceResponse {
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
		CreatedAt:             instance.CreatedAt,
		UpdatedAt:             instance.UpdatedAt,
	}
}

// GetResourceInstanceDTOsFromDBModels maps stored instances onto their API responses.
// The slice is never nil, so the endpoint answers with [] rather than null.
func GetResourceInstanceDTOsFromDBModels(instances []db.ResourceInstance) []ResourceInstanceResponse {
	responses := make([]ResourceInstanceResponse, 0, len(instances))
	for _, instance := range instances {
		responses = append(responses, GetResourceInstanceDTOFromDBModel(instance))
	}
	return responses
}
