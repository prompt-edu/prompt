package seatPlanDTO

import (
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/intro_course/db/sqlc"
)

type SeatAssignment struct {
	SeatName                     string      `json:"seatName"`
	HasMac                       bool        `json:"hasMac"`
	DeviceID                     pgtype.Text `json:"deviceID"`
	StudentCourseParticipationID pgtype.UUID `json:"studentCourseParticipationID"`
	TutorFirstName               string      `json:"tutorFirstName"`
	TutorLastName                string      `json:"tutorLastName"`
	TutorEmail                   string      `json:"tutorEmail"`
}

func GetSeatAssignmentDTOFromDBModel(seatAssignment db.GetOwnSeatAssignmentRow) SeatAssignment {
	return SeatAssignment{
		SeatName:                     seatAssignment.SeatName,
		HasMac:                       seatAssignment.HasMac,
		DeviceID:                     seatAssignment.DeviceID,
		StudentCourseParticipationID: seatAssignment.AssignedStudent,
		TutorFirstName:               seatAssignment.TutorFirstName,
		TutorLastName:                seatAssignment.TutorLastName,
		TutorEmail:                   seatAssignment.TutorEmail,
	}
}
