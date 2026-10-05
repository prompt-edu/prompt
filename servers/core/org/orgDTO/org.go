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
	ParentOrgID  *uuid.UUID  `json:"parentOrgID"`
	Name         string      `json:"name"`
	Slug         string      `json:"slug"`
	School       pgtype.Text `json:"school" swaggertype:"string"`
	University   pgtype.Text `json:"university" swaggertype:"string"`
	Website      pgtype.Text `json:"website" swaggertype:"string"`
	ContactEmail pgtype.Text `json:"contactEmail" swaggertype:"string"`
	CreatedAt    time.Time   `json:"createdAt"`
}

// CreateOrg creates an org, optionally under a parent org. Empty optional fields are
// stored as null.
type CreateOrg struct {
	ParentOrgID  *uuid.UUID `json:"parentOrgID"`
	Name         string     `json:"name"`
	Slug         string     `json:"slug"`
	School       string     `json:"school"`
	University   string     `json:"university"`
	Website      string     `json:"website"`
	ContactEmail string     `json:"contactEmail"`
}

// UpdateOrg replaces the name and all optional fields; an omitted, null, or empty
// optional field clears it. The slug and the parent org cannot be changed here.
type UpdateOrg struct {
	Name         string `json:"name"`
	School       string `json:"school"`
	University   string `json:"university"`
	Website      string `json:"website"`
	ContactEmail string `json:"contactEmail"`
}

// UpdateOrgParent moves an org under another org, or to the top level when
// ParentOrgID is null.
type UpdateOrgParent struct {
	ParentOrgID *uuid.UUID `json:"parentOrgID"`
}

// UnmarshalJSON requires the parentOrgID key, so that a body without it, such as {} or a
// misspelled key, is rejected instead of moving the org to the top level.
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
