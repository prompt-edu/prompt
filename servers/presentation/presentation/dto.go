package presentation

import (
	"time"

	"github.com/google/uuid"
)

type SettingsResponse struct {
	CoursePhaseID uuid.UUID `json:"coursePhaseId"`
	TargetMode    string    `json:"targetMode"`
	FeedbackMode  string    `json:"feedbackMode"`
}

type UpdateSettingsRequest struct {
	TargetMode        string `json:"targetMode" binding:"required,oneof=individual team"`
	FeedbackMode      string `json:"feedbackMode" binding:"required,oneof=independent shared"`
	ResetExistingData bool   `json:"resetExistingData"`
}

type CategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Position    int32  `json:"position" binding:"min=0"`
}

type CategoryResponse struct {
	ID            uuid.UUID `json:"id"`
	CoursePhaseID uuid.UUID `json:"coursePhaseId"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Position      int32     `json:"position"`
}

type SlotRequest struct {
	StartTime time.Time `json:"startTime" binding:"required"`
	EndTime   time.Time `json:"endTime" binding:"required"`
	Location  string    `json:"location"`
}

type SlotResponse struct {
	ID            uuid.UUID             `json:"id"`
	CoursePhaseID uuid.UUID             `json:"coursePhaseId"`
	StartTime     time.Time             `json:"startTime"`
	EndTime       time.Time             `json:"endTime"`
	Location      string                `json:"location"`
	Presentation  *PresentationResponse `json:"presentation,omitempty"`
}

type TargetResponse struct {
	ID                     uuid.UUID  `json:"id"`
	Type                   string     `json:"type"`
	Name                   string     `json:"name"`
	AssignedPresentationID *uuid.UUID `json:"assignedPresentationId,omitempty"`
}

type AssignmentRequest struct {
	TargetID   uuid.UUID `json:"targetId" binding:"required"`
	TargetName string    `json:"targetName" binding:"required"`
	TargetType string    `json:"targetType" binding:"required,oneof=individual team"`
}

type PresentationResponse struct {
	ID                     uuid.UUID  `json:"id"`
	CoursePhaseID          uuid.UUID  `json:"coursePhaseId"`
	SlotID                 uuid.UUID  `json:"slotId"`
	TargetType             string     `json:"targetType"`
	TargetID               uuid.UUID  `json:"targetId"`
	TargetName             string     `json:"targetName"`
	StartTime              time.Time  `json:"startTime"`
	EndTime                time.Time  `json:"endTime"`
	Location               string     `json:"location"`
	MaterialCount          int64      `json:"materialCount"`
	FeedbackCount          int64      `json:"feedbackCount"`
	SubmittedFeedbackCount int64      `json:"submittedFeedbackCount"`
	FeedbackReleaseName    *string    `json:"feedbackReleaseName,omitempty"`
	FeedbackReleasedAt     *time.Time `json:"feedbackReleasedAt,omitempty"`
	FeedbackReleasedByName *string    `json:"feedbackReleasedByName,omitempty"`
}

type PresignMaterialRequest struct {
	FileName    string `json:"fileName" binding:"required"`
	ContentType string `json:"contentType" binding:"required"`
	SizeBytes   int64  `json:"sizeBytes" binding:"required,min=1"`
}

type PresignMaterialResponse struct {
	UploadID  uuid.UUID         `json:"uploadId"`
	UploadURL string            `json:"uploadUrl"`
	Headers   map[string]string `json:"headers,omitempty"`
	ExpiresAt time.Time         `json:"expiresAt"`
}

type MaterialResponse struct {
	ID             uuid.UUID `json:"id"`
	PresentationID uuid.UUID `json:"presentationId"`
	FileName       string    `json:"fileName"`
	ContentType    string    `json:"contentType"`
	SizeBytes      int64     `json:"sizeBytes"`
	UploadedByName string    `json:"uploadedByName"`
	UploadedAt     time.Time `json:"uploadedAt"`
}

type MaterialDownloadResponse struct {
	DownloadURL string    `json:"downloadUrl"`
	FileName    string    `json:"fileName"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type PutAnswerRequest struct {
	Value            string `json:"value"`
	ExpectedRevision int64  `json:"expectedRevision" binding:"min=0"`
}

type ReleaseFeedbackRequest struct {
	ReleaseName string `json:"releaseName" binding:"required"`
}

type FeedbackAnswerResponse struct {
	CategoryID    uuid.UUID `json:"categoryId"`
	Value         string    `json:"value"`
	Revision      int64     `json:"revision"`
	UpdatedByName string    `json:"updatedByName,omitempty"`
	UpdatedAt     time.Time `json:"updatedAt,omitempty"`
}

type ContributorResponse struct {
	UserID             string    `json:"userId"`
	Name               string    `json:"name"`
	FirstContributedAt time.Time `json:"firstContributedAt"`
	LastContributedAt  time.Time `json:"lastContributedAt"`
}

type FeedbackFormResponse struct {
	ID            uuid.UUID                `json:"id"`
	EvaluatorName string                   `json:"evaluatorName,omitempty"`
	Status        string                   `json:"status"`
	SubmittedAt   *time.Time               `json:"submittedAt,omitempty"`
	Answers       []FeedbackAnswerResponse `json:"answers"`
	Contributors  []ContributorResponse    `json:"contributors"`
	IsOwn         bool                     `json:"isOwn"`
}

type FeedbackDocumentResponse struct {
	Presentation  PresentationResponse   `json:"presentation"`
	Mode          string                 `json:"mode"`
	Categories    []CategoryResponse     `json:"categories"`
	OwnForm       *FeedbackFormResponse  `json:"ownForm,omitempty"`
	Forms         []FeedbackFormResponse `json:"forms"`
	ActiveEditors []ActiveEditorResponse `json:"activeEditors"`
	CanEdit       bool                   `json:"canEdit"`
	CanRelease    bool                   `json:"canRelease"`
}

type ActiveEditorResponse struct {
	ConnectionID uuid.UUID `json:"connectionId"`
	UserID       string    `json:"userId"`
	Name         string    `json:"name"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

type FeedbackEvent struct {
	ID             string                  `json:"id,omitempty"`
	Type           string                  `json:"type"`
	PresentationID uuid.UUID               `json:"presentationId"`
	Answer         *FeedbackAnswerResponse `json:"answer,omitempty"`
	ActiveEditors  []ActiveEditorResponse  `json:"activeEditors,omitempty"`
}
