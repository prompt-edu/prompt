package seatPlanDTO

import (
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/intro_course/db/sqlc"
)

type Seat struct {
	SeatName        string      `json:"seatName"`
	HasMac          bool        `json:"hasMac"`
	DeviceID        pgtype.Text `json:"deviceID"`
	AssignedStudent pgtype.UUID `json:"assignedStudent"` // using pgtype bc. it might be empty
	AssignedTutor   pgtype.UUID `json:"assignedTutor"`
}

func GetSeatDTOFromDBModel(seat db.Seat) Seat {
	return Seat{
		SeatName:        seat.SeatName,
		HasMac:          seat.HasMac,
		DeviceID:        seat.DeviceID,
		AssignedStudent: seat.AssignedStudent,
		AssignedTutor:   seat.AssignedTutor,
	}
}

func GetSeatDTOsFromDBModels(seats []db.Seat) []Seat {
	seatDTOs := make([]Seat, 0, len(seats))
	for _, seat := range seats {
		seatDTOs = append(seatDTOs, GetSeatDTOFromDBModel(seat))
	}
	return seatDTOs
}
