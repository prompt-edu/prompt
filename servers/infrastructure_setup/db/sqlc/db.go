package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX is satisfied by both *pgxpool.Pool and pgx.Tx.
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

// Queries holds the database connection and exposes all query methods.
type Queries struct {
	db DBTX
}

// New creates a Queries instance backed by a connection pool.
func New(db *pgxpool.Pool) *Queries {
	return &Queries{db: db}
}

// WithTx returns a Queries instance scoped to a transaction.
func (q *Queries) WithTx(tx pgx.Tx) *Queries {
	return &Queries{db: tx}
}

// ProviderType is the Go representation of the provider_type SQL enum.
type ProviderType string

const (
	ProviderTypeGitlab   ProviderType = "gitlab"
	ProviderTypeSlack    ProviderType = "slack"
	ProviderTypeOutline  ProviderType = "outline"
	ProviderTypeRancher  ProviderType = "rancher"
	ProviderTypeKeycloak ProviderType = "keycloak"
)

// ResourceScope is the Go representation of the resource_scope SQL enum.
type ResourceScope string

const (
	ResourceScopePerTeam    ResourceScope = "per_team"
	ResourceScopePerStudent ResourceScope = "per_student"
)

// ResourceStatus is the Go representation of the resource_status SQL enum.
type ResourceStatus string

const (
	ResourceStatusPending    ResourceStatus = "pending"
	ResourceStatusInProgress ResourceStatus = "in_progress"
	ResourceStatusCreated    ResourceStatus = "created"
	ResourceStatusFailed     ResourceStatus = "failed"
)

// ProviderConfig is the row type for the provider_config table.
type ProviderConfig struct {
	ID             uuid.UUID    `json:"id"`
	CoursePhaseID  uuid.UUID    `json:"coursePhaseId"`
	ProviderType   ProviderType `json:"providerType"`
	Credentials    []byte       `json:"-"` // never serialized in API responses
}

// ResourceConfig is the row type for the resource_config table.
type ResourceConfig struct {
	ID                  uuid.UUID       `json:"id"`
	CoursePhaseID       uuid.UUID       `json:"coursePhaseId"`
	ProviderType        ProviderType    `json:"providerType"`
	ResourceType        string          `json:"resourceType"`
	Scope               ResourceScope   `json:"scope"`
	NameTemplate        string          `json:"nameTemplate"`
	PermissionMapping   json.RawMessage `json:"permissionMapping"`
	ResourceExtraConfig json.RawMessage `json:"resourceExtraConfig"`
	CreatedAt           time.Time       `json:"createdAt"`
}

// ResourceInstance is the row type for the resource_instance table.
type ResourceInstance struct {
	ID                    uuid.UUID      `json:"id"`
	ResourceConfigID      uuid.UUID      `json:"resourceConfigId"`
	CoursePhaseID         uuid.UUID      `json:"coursePhaseId"`
	TeamID                *uuid.UUID     `json:"teamId,omitempty"`
	CourseParticipationID *uuid.UUID     `json:"courseParticipationId,omitempty"`
	Status                ResourceStatus `json:"status"`
	ExternalID            *string        `json:"externalId,omitempty"`
	ExternalURL           *string        `json:"externalUrl,omitempty"`
	ErrorMessage          *string        `json:"errorMessage,omitempty"`
	RetryCount            int32          `json:"retryCount"`
	CreatedAt             time.Time      `json:"createdAt"`
	UpdatedAt             time.Time      `json:"updatedAt"`
}
