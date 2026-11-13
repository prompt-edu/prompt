package orgDTO

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type Org struct {
	ID           uuid.UUID   `json:"id"`
	ParentOrgID  *uuid.UUID  `json:"parentOrgID" extensions:"x-nullable"`
	Name         string      `json:"name"`
	Slug         string      `json:"slug"`
	School       pgtype.Text `json:"school" swaggertype:"string" extensions:"x-nullable"`
	University   pgtype.Text `json:"university" swaggertype:"string" extensions:"x-nullable"`
	Website      pgtype.Text `json:"website" swaggertype:"string" extensions:"x-nullable"`
	ContactEmail pgtype.Text `json:"contactEmail" swaggertype:"string" extensions:"x-nullable"`
	CreatedAt    time.Time   `json:"createdAt"`
}

type CreateOrg struct {
	ParentOrgID  *uuid.UUID `json:"parentOrgID" extensions:"x-nullable"`
	Name         string     `json:"name" validate:"required"`
	Slug         string     `json:"slug" validate:"required"`
	School       string     `json:"school"`
	University   string     `json:"university"`
	Website      string     `json:"website"`
	ContactEmail string     `json:"contactEmail"`
}

type UpdateOrg struct {
	Name         string `json:"name" validate:"required"`
	School       string `json:"school"`
	University   string `json:"university"`
	Website      string `json:"website"`
	ContactEmail string `json:"contactEmail"`
}

// OwnOrg is an org whose role the caller's token holds.
type OwnOrg struct {
	Org
	// Role is the caller's org role, Admin or Member.
	Role string `json:"role" enums:"Admin,Member"`
}

// SetOrgMember gives the user with the university login an org role, adding them to the org if
// they are not a member yet.
type SetOrgMember struct {
	UniversityLogin string `json:"universityLogin" validate:"required"`
	Role            string `json:"role" validate:"required" enums:"Admin,Member"`
}

type UpdateOrgParent struct {
	ParentOrgID *uuid.UUID `json:"parentOrgID" validate:"required" extensions:"x-nullable"`
}

func (u *UpdateOrgParent) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	parentOrgID, ok := fields["parentOrgID"]
	if !ok {
		return errors.New("parentOrgID is required; pass null to move the org to the top level")
	}
	return json.Unmarshal(parentOrgID, &u.ParentOrgID)
}

func GetOrgDTOFromDBModel(model db.Org) Org {
	var parentOrgID *uuid.UUID
	if model.ParentOrgID.Valid {
		id := uuid.UUID(model.ParentOrgID.Bytes)
		parentOrgID = &id
	}
	return Org{
		ID:           model.ID,
		ParentOrgID:  parentOrgID,
		Name:         model.Name,
		Slug:         model.Slug,
		School:       model.School,
		University:   model.University,
		Website:      model.Website,
		ContactEmail: model.ContactEmail,
		CreatedAt:    model.CreatedAt.Time,
	}
}

func GetOrgDTOsFromDBModels(models []db.Org) []Org {
	orgs := make([]Org, 0, len(models))
	for _, model := range models {
		orgs = append(orgs, GetOrgDTOFromDBModel(model))
	}
	return orgs
}
