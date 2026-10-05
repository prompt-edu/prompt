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
	// Members lists the people the instance's latest run was for, and whether each was
	// granted access. Empty until the instance has run.
	Members   []InstanceMemberResponse `json:"members"`
	CreatedAt time.Time                `json:"createdAt"`
	UpdatedAt time.Time                `json:"updatedAt"`
}

// InstanceMemberResponse is one person an instance's latest run was for.
type InstanceMemberResponse struct {
	CourseParticipationID uuid.UUID `json:"courseParticipationId"`
	Granted               bool      `json:"granted"`
}

// GetResourceInstanceDTOFromDBModel maps one listed instance and its members onto its
// API response.
func GetResourceInstanceDTOFromDBModel(instance db.ListResourceInstancesWithConfigRow, members []InstanceMemberResponse) ResourceInstanceResponse {
	if members == nil {
		members = []InstanceMemberResponse{}
	}
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
		Members:               members,
		CreatedAt:             instance.CreatedAt,
		UpdatedAt:             instance.UpdatedAt,
	}
}

// GetResourceInstanceDTOsFromDBModels maps listed instances onto their API responses,
// each carrying its own members. The slice is never nil, so the endpoint answers with []
// rather than null.
func GetResourceInstanceDTOsFromDBModels(instances []db.ListResourceInstancesWithConfigRow, members []db.ResourceInstanceMember) []ResourceInstanceResponse {
	membersByInstance := make(map[uuid.UUID][]InstanceMemberResponse, len(instances))
	for _, member := range members {
		membersByInstance[member.ResourceInstanceID] = append(membersByInstance[member.ResourceInstanceID], InstanceMemberResponse{
			CourseParticipationID: member.CourseParticipationID,
			Granted:               member.Granted,
		})
	}

	responses := make([]ResourceInstanceResponse, 0, len(instances))
	for _, instance := range instances {
		responses = append(responses, GetResourceInstanceDTOFromDBModel(instance, membersByInstance[instance.ID]))
	}
	return responses
}

// MyResourceResponse is one configured resource as a student of the phase sees it.
//
// It deliberately carries no error text: a partial run's message names the other members
// who could not be added, and a failure's message is about the course's credentials.
type MyResourceResponse struct {
	ResourceConfigID uuid.UUID        `json:"resourceConfigId"`
	ProviderType     db.ProviderType  `json:"providerType"`
	ResourceType     string           `json:"resourceType"`
	Scope            db.ResourceScope `json:"scope"`
	// Status is the status of the student's instance, or null when nothing has been
	// provisioned for them from this config yet.
	Status *db.ResourceStatus `json:"status"`
	// Granted says whether the latest run let the student in. It is null when there is no
	// instance, or when its run predates member tracking.
	Granted *bool `json:"granted"`
	// Name is the name of the resource the provider was asked to create.
	Name string `json:"name"`
	// TeamName names the team a per_team resource belongs to.
	TeamName string `json:"teamName"`
	// URL links to the resource, when there is anything there a student can open.
	URL *string `json:"url"`
}

// ProvisioningPreview says what a trigger would do right now, without doing it.
type ProvisioningPreview struct {
	// Queued counts targets that have no instance yet.
	Queued int `json:"queued"`
	// Requeued counts failed or partial instances a trigger would retry.
	Requeued int `json:"requeued"`
	// UpToDate counts targets whose resource is already created.
	UpToDate int `json:"upToDate"`
	// Running counts instances a run is still working on. While it is above zero a
	// trigger is refused.
	Running int `json:"running"`
	// Teams counts the teams resolved, or is null when no resource config is per_team.
	Teams *int `json:"teams"`
	// Students counts the students resolved, or is null when no resource config is
	// per_student.
	Students *int `json:"students"`
}
