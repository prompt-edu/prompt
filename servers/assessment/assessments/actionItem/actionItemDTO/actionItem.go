package actionItemDTO

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type ActionItem struct {
	ID                    uuid.UUID        `json:"id"`
	CoursePhaseID         uuid.UUID        `json:"coursePhaseID"`
	CourseParticipationID uuid.UUID        `json:"courseParticipationID"`
	Action                string           `json:"action"`
	CreatedAt             pgtype.Timestamp `json:"createdAt"`
	Author                string           `json:"author"`
}

// GetActionItemDTOsFromDBModels converts a slice of db.ActionItem to DTOs.
func GetActionItemDTOsFromDBModels(dbActionItems []db.ActionItem) []ActionItem {
	actionItems := make([]ActionItem, 0, len(dbActionItems))
	for _, a := range dbActionItems {
		actionItems = append(actionItems, MapDBActionItemToActionItemDTO(a))
	}
	return actionItems
}

// MapDBActionItemToActionItemDTO converts a db.ActionItem to DTO.
func MapDBActionItemToActionItemDTO(dbActionItem db.ActionItem) ActionItem {
	return ActionItem{
		ID:                    dbActionItem.ID,
		CoursePhaseID:         dbActionItem.CoursePhaseID,
		CourseParticipationID: dbActionItem.CourseParticipationID,
		Action:                dbActionItem.Action,
		CreatedAt:             dbActionItem.CreatedAt,
		Author:                dbActionItem.Author,
	}
}

// GetDBModel converts CreateActionItemRequest to database parameters.
func (r CreateActionItemRequest) GetDBModel() db.CreateActionItemParams {
	return db.CreateActionItemParams{
		ID:                    uuid.New(),
		CoursePhaseID:         r.CoursePhaseID,
		CourseParticipationID: r.CourseParticipationID,
		Action:                r.Action,
		Author:                r.Author,
	}
}

// GetDBModel converts UpdateActionItemRequest to database parameters.
func (r UpdateActionItemRequest) GetDBModel() db.UpdateActionItemParams {
	return db.UpdateActionItemParams{
		ID:                    r.ID,
		CoursePhaseID:         r.CoursePhaseID,
		CourseParticipationID: r.CourseParticipationID,
		Action:                r.Action,
		Author:                r.Author,
	}
}
