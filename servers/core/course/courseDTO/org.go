package courseDTO

import (
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// UpdateCourseOrg assigns a course to an org, or clears its org with a null OrgID.
type UpdateCourseOrg struct {
	OrgID *uuid.UUID `json:"orgID" validate:"required" extensions:"x-nullable"`
}

func (u *UpdateCourseOrg) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	orgID, ok := fields["orgID"]
	if !ok {
		return errors.New("orgID is required; pass null to remove the course from its org")
	}
	return json.Unmarshal(orgID, &u.OrgID)
}

func OrgIDToDBModel(orgID *uuid.UUID) pgtype.UUID {
	if orgID == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *orgID, Valid: true}
}

func OrgIDFromDBModel(orgID pgtype.UUID) *uuid.UUID {
	if !orgID.Valid {
		return nil
	}
	id := uuid.UUID(orgID.Bytes)
	return &id
}
