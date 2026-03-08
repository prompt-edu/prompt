package configDTO

import (
	"time"

	"github.com/google/uuid"
	db "github.com/ls1intum/prompt2/servers/certificate/db/sqlc"
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
}

type UpdateConfigRequest struct {
	TemplateContent string `json:"templateContent" binding:"required"`
}

type UpdateReleaseDateRequest struct {
	ReleaseDate *time.Time `json:"releaseDate"`
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

	return CoursePhaseConfig{
		CoursePhaseID:   dbConfig.CoursePhaseID,
		TemplateContent: templateContent,
		HasTemplate:     dbConfig.TemplateContent.Valid && dbConfig.TemplateContent.String != "",
		CreatedAt:       dbConfig.CreatedAt.Time,
		UpdatedAt:       dbConfig.UpdatedAt.Time,
		UpdatedBy:       updatedBy,
		ReleaseDate:     releaseDate,
		HasDownloads:    hasDownloads,
	}
}
