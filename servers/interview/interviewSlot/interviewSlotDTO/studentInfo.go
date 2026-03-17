package interviewSlotDTO

import "github.com/google/uuid"

type StudentInfo struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
}
