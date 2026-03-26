package participants

import (
	"time"

	"github.com/google/uuid"
)

type Student struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	MatrNr    *string   `json:"matrNr,omitempty"`
}

type CoursePhaseParticipation struct {
	CoursePhaseID         uuid.UUID              `json:"coursePhaseID"`
	CourseParticipationID uuid.UUID              `json:"courseParticipationID"`
	PassStatus            string                 `json:"passStatus"`
	RestrictedData        map[string]interface{} `json:"restrictedData"`
	StudentReadableData   map[string]interface{} `json:"studentReadableData"`
	PrevData              map[string]interface{} `json:"prevData"`
	Student               Student                `json:"student"`
}

type CoursePhaseParticipationsResponse struct {
	CoursePhaseParticipations []CoursePhaseParticipation `json:"participations"`
}

// ParticipantWithDownloadStatus is the enriched response returned by the participants endpoint.
// It extends CoursePhaseParticipation with download-tracking fields.
type ParticipantWithDownloadStatus struct {
	CoursePhaseParticipation
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
	CoursePhaseID         uuid.UUID `json:"coursePhaseID"`
	CourseParticipationID uuid.UUID `json:"courseParticipationID"`
	Student               Student   `json:"student"`
}
