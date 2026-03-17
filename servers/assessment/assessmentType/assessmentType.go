package assessmentType

import (
	"log"

	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type AssessmentType string

const (
	Self       AssessmentType = "self"
	Peer       AssessmentType = "peer"
	Tutor      AssessmentType = "tutor"
	Assessment AssessmentType = "assessment"
)

func MapDBAssessmentTypeToDTO(assessmentType db.AssessmentType) AssessmentType {
	switch assessmentType {
	case db.AssessmentTypeSelf:
		return Self
	case db.AssessmentTypePeer:
		return Peer
	case db.AssessmentTypeTutor:
		return Tutor
	case db.AssessmentTypeAssessment:
		return Assessment
	}

	log.Println("Warning: Unrecognized assessment type in MapDBAssessmentTypeToDTO:", assessmentType)
	return Self // Default case, should not happen
}

func MapDTOtoDBAssessmentType(assessmentType AssessmentType) db.AssessmentType {
	switch assessmentType {
	case Self:
		return db.AssessmentTypeSelf
	case Peer:
		return db.AssessmentTypePeer
	case Tutor:
		return db.AssessmentTypeTutor
	case Assessment:
		return db.AssessmentTypeAssessment
	}

	log.Println("Warning: Unrecognized assessment type in MapDTOtoDBAssessmentType:", assessmentType)
	return db.AssessmentTypeSelf // Default case, should not happen
}
