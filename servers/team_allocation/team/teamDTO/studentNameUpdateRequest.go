package teamDTO

import "github.com/google/uuid"

type StudentNameUpdateRequest struct {
	StudentNamesPerID map[uuid.UUID]StudentName `json:"studentNamesPerID"`
}

type StudentName struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}
