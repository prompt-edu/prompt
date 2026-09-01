package configDTO

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/certificate/db/sqlc"
)

// MaxStudentPageTextBytes caps the instructor text rendered on the student
// download page.
const MaxStudentPageTextBytes = 20000

var (
	ErrStudentPageTextMissing = errors.New("studentPageText is required")
	ErrStudentPageTextType    = errors.New("studentPageText must be a string or null")
	ErrStudentPageTextTooLong = errors.New("studentPageText is too long")
)

type CoursePhaseConfig struct {
	CoursePhaseID   uuid.UUID  `json:"coursePhaseId"`
	TemplateContent *string    `json:"templateContent,omitempty"`
	HasTemplate     bool       `json:"hasTemplate"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	UpdatedBy       *string    `json:"updatedBy,omitempty"`
	ReleaseDate     *time.Time `json:"releaseDate,omitempty"`
	HasDownloads    bool       `json:"hasDownloads"`
	StudentPageText *string    `json:"studentPageText,omitempty"`
}

type UpdateConfigRequest struct {
	TemplateContent string `json:"templateContent" binding:"required"`
}

type UpdateReleaseDateRequest struct {
	ReleaseDate *time.Time `json:"releaseDate"`
}

// UpdateStudentPageTextRequest keeps the raw value so an omitted key can be
// told apart from an explicit null, which clears the text.
type UpdateStudentPageTextRequest struct {
	StudentPageText json.RawMessage `json:"studentPageText"`
}

// Text returns the requested value: nil clears the stored text.
func (r UpdateStudentPageTextRequest) Text() (*string, error) {
	if len(r.StudentPageText) == 0 {
		return nil, ErrStudentPageTextMissing
	}
	if string(r.StudentPageText) == "null" {
		return nil, nil
	}

	var text string
	if err := json.Unmarshal(r.StudentPageText, &text); err != nil {
		return nil, ErrStudentPageTextType
	}
	if len(text) > MaxStudentPageTextBytes {
		return nil, ErrStudentPageTextTooLong
	}
	return &text, nil
}

func MapDBConfigToDTOConfig(dbConfig db.CoursePhaseConfig, hasDownloads bool) CoursePhaseConfig {
	var templateContent *string
	if dbConfig.TemplateContent.Valid {
		templateContent = &dbConfig.TemplateContent.String
	}

	var updatedBy *string
	if dbConfig.UpdatedBy.Valid {
		updatedBy = &dbConfig.UpdatedBy.String
	}

	var releaseDate *time.Time
	if dbConfig.ReleaseDate.Valid {
		t := dbConfig.ReleaseDate.Time
		releaseDate = &t
	}

	var studentPageText *string
	if dbConfig.StudentPageText.Valid && dbConfig.StudentPageText.String != "" {
		studentPageText = &dbConfig.StudentPageText.String
	}

	return CoursePhaseConfig{
		CoursePhaseID:   dbConfig.CoursePhaseID,
		TemplateContent: templateContent,
		HasTemplate:     dbConfig.TemplateContent.Valid && dbConfig.TemplateContent.String != "",
		CreatedAt:       dbConfig.CreatedAt.Time,
		UpdatedAt:       dbConfig.UpdatedAt.Time,
		UpdatedBy:       updatedBy,
		ReleaseDate:     releaseDate,
		HasDownloads:    hasDownloads,
		StudentPageText: studentPageText,
	}
}
