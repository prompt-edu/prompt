package callDTO

import (
	"time"

	"github.com/google/uuid"
)

type Cursor struct {
	RequestedAt time.Time `json:"requestedAt"`
	ID          uuid.UUID `json:"id"`
}

type Page struct {
	Calls      []Call  `json:"calls"`
	NextCursor *Cursor `json:"nextCursor"`
}
