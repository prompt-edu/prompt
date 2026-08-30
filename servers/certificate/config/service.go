package config

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prompt-edu/prompt/servers/certificate/config/configDTO"
	db "github.com/prompt-edu/prompt/servers/certificate/db/sqlc"
	log "github.com/sirupsen/logrus"
)

type ConfigService struct {
	queries db.Queries
}

func NewConfigService(queries db.Queries) *ConfigService {
	return &ConfigService{
		queries: queries,
	}
}

func (s *ConfigService) GetCoursePhaseConfig(ctx context.Context, coursePhaseID uuid.UUID) (configDTO.CoursePhaseConfig, error) {
	config, err := s.queries.GetCoursePhaseConfig(ctx, coursePhaseID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Create a default config
			config, err = s.queries.CreateCoursePhaseConfig(ctx, coursePhaseID)
			if err != nil {
				log.WithError(err).Error("Failed to create default course phase config")
				return configDTO.CoursePhaseConfig{}, err
			}
		} else {
			log.WithError(err).Error("Failed to get course phase config")
			return configDTO.CoursePhaseConfig{}, err
		}
	}

	hasDownloads, err := s.queries.HasDownloads(ctx, coursePhaseID)
	if err != nil {
		log.WithError(err).Warn("Failed to check for existing downloads")
		hasDownloads = false
	}

	return configDTO.MapDBConfigToDTOConfig(config, hasDownloads), nil
}

func (s *ConfigService) UpdateCoursePhaseConfig(ctx context.Context, coursePhaseID uuid.UUID, templateContent string, updatedBy string) (configDTO.CoursePhaseConfig, error) {
	config, err := s.queries.UpsertCoursePhaseConfig(ctx, db.UpsertCoursePhaseConfigParams{
		CoursePhaseID:   coursePhaseID,
		TemplateContent: pgtype.Text{String: templateContent, Valid: true},
		UpdatedBy:       pgtype.Text{String: updatedBy, Valid: updatedBy != ""},
	})
	if err != nil {
		log.WithError(err).Error("Failed to update course phase config")
		return configDTO.CoursePhaseConfig{}, err
	}

	hasDownloads, err := s.queries.HasDownloads(ctx, coursePhaseID)
	if err != nil {
		log.WithError(err).Warn("Failed to check for existing downloads")
		hasDownloads = false
	}

	return configDTO.MapDBConfigToDTOConfig(config, hasDownloads), nil
}

func (s *ConfigService) UpdateReleaseDate(ctx context.Context, coursePhaseID uuid.UUID, releaseDate *time.Time, updatedBy string) (configDTO.CoursePhaseConfig, error) {
	var releaseDatePg pgtype.Timestamptz
	if releaseDate != nil {
		releaseDatePg = pgtype.Timestamptz{Time: *releaseDate, Valid: true}
	} else {
		releaseDatePg = pgtype.Timestamptz{Valid: false}
	}

	config, err := s.queries.UpdateReleaseDate(ctx, db.UpdateReleaseDateParams{
		CoursePhaseID: coursePhaseID,
		ReleaseDate:   releaseDatePg,
		UpdatedBy:     pgtype.Text{String: updatedBy, Valid: updatedBy != ""},
	})
	if err != nil {
		log.WithError(err).Error("Failed to update release date")
		return configDTO.CoursePhaseConfig{}, err
	}

	hasDownloads, err := s.queries.HasDownloads(ctx, coursePhaseID)
	if err != nil {
		log.WithError(err).Warn("Failed to check for existing downloads")
		hasDownloads = false
	}

	return configDTO.MapDBConfigToDTOConfig(config, hasDownloads), nil
}

func (s *ConfigService) UpdateStudentPageText(ctx context.Context, coursePhaseID uuid.UUID, studentPageText *string) (configDTO.CoursePhaseConfig, error) {
	var text pgtype.Text
	if studentPageText != nil {
		text = pgtype.Text{String: *studentPageText, Valid: true}
	}

	config, err := s.queries.UpsertStudentPageText(ctx, db.UpsertStudentPageTextParams{
		CoursePhaseID:   coursePhaseID,
		StudentPageText: text,
	})
	if err != nil {
		log.WithError(err).Error("Failed to update student page text")
		return configDTO.CoursePhaseConfig{}, err
	}

	hasDownloads, err := s.queries.HasDownloads(ctx, coursePhaseID)
	if err != nil {
		log.WithError(err).Warn("Failed to check for existing downloads")
		hasDownloads = false
	}

	return configDTO.MapDBConfigToDTOConfig(config, hasDownloads), nil
}

func (s *ConfigService) GetTemplateContent(ctx context.Context, coursePhaseID uuid.UUID) (string, error) {
	config, err := s.queries.GetCoursePhaseConfig(ctx, coursePhaseID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New("no template configured for this course phase")
		}
		log.WithError(err).Error("Failed to get course phase config")
		return "", err
	}

	if !config.TemplateContent.Valid || config.TemplateContent.String == "" {
		return "", errors.New("no template configured for this course phase")
	}

	return config.TemplateContent.String, nil
}
