package participants

import (
	"time"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
)

type CoursePhaseParticipationsResponse struct {
	CoursePhaseParticipations []promptTypes.CoursePhaseParticipationWithStudent `json:"participations"`
}

// ParticipantWithDownloadStatus is the enriched response returned by the participants endpoint.
// It extends CoursePhaseParticipationWithStudent with download-tracking fields.
type ParticipantWithDownloadStatus struct {
	promptTypes.CoursePhaseParticipationWithStudent
	HasDownloaded bool       `json:"hasDownloaded"`
	FirstDownload *time.Time `json:"firstDownload,omitempty"`
	LastDownload  *time.Time `json:"lastDownload,omitempty"`
	DownloadCount int32      `json:"downloadCount"`
}

type Course struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type CoursePhaseWithCourse struct {
	ID       uuid.UUID `json:"id"`
	CourseID uuid.UUID `json:"courseId"`
	Course   Course    `json:"course"`
}

// CoursePhaseParticipationSelf is the response from core's /participations/self endpoint
type CoursePhaseParticipationSelf struct {
	CoursePhaseID         uuid.UUID           `json:"coursePhaseID"`
	CourseParticipationID uuid.UUID           `json:"courseParticipationID"`
	Student               promptTypes.Student `json:"student"`
}
