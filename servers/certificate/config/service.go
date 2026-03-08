package config

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ls1intum/prompt2/servers/certificate/config/configDTO"
	db "github.com/ls1intum/prompt2/servers/certificate/db/sqlc"
	log "github.com/sirupsen/logrus"
)

type ConfigService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var ConfigServiceSingleton *ConfigService

func NewConfigService(queries db.Queries, conn *pgxpool.Pool) *ConfigService {
	return &ConfigService{
		queries: queries,
		conn:    conn,
	}
}

func GetCoursePhaseConfig(ctx context.Context, coursePhaseID uuid.UUID) (configDTO.CoursePhaseConfig, error) {
	config, err := ConfigServiceSingleton.queries.GetCoursePhaseConfig(ctx, coursePhaseID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Create a default config
			config, err = ConfigServiceSingleton.queries.CreateCoursePhaseConfig(ctx, coursePhaseID)
			if err != nil {
				log.WithError(err).Error("Failed to create default course phase config")
				return configDTO.CoursePhaseConfig{}, err
			}
		} else {
			log.WithError(err).Error("Failed to get course phase config")
			return configDTO.CoursePhaseConfig{}, err
		}
	}

	hasDownloads, err := ConfigServiceSingleton.queries.HasDownloads(ctx, coursePhaseID)
	if err != nil {
		log.WithError(err).Warn("Failed to check for existing downloads")
		hasDownloads = false
	}

	return configDTO.MapDBConfigToDTOConfig(config, hasDownloads), nil
}

func UpdateCoursePhaseConfig(ctx context.Context, coursePhaseID uuid.UUID, templateContent string, updatedBy string) (configDTO.CoursePhaseConfig, error) {
	config, err := ConfigServiceSingleton.queries.UpsertCoursePhaseConfig(ctx, db.UpsertCoursePhaseConfigParams{
		CoursePhaseID:   coursePhaseID,
		TemplateContent: pgtype.Text{String: templateContent, Valid: true},
		UpdatedBy:       pgtype.Text{String: updatedBy, Valid: updatedBy != ""},
	})
	if err != nil {
		log.WithError(err).Error("Failed to update course phase config")
		return configDTO.CoursePhaseConfig{}, err
	}

	hasDownloads, err := ConfigServiceSingleton.queries.HasDownloads(ctx, coursePhaseID)
	if err != nil {
		log.WithError(err).Warn("Failed to check for existing downloads")
		hasDownloads = false
	}

	return configDTO.MapDBConfigToDTOConfig(config, hasDownloads), nil
}

func UpdateReleaseDate(ctx context.Context, coursePhaseID uuid.UUID, releaseDate *time.Time, updatedBy string) (configDTO.CoursePhaseConfig, error) {
	var releaseDatePg pgtype.Timestamptz
	if releaseDate != nil {
		releaseDatePg = pgtype.Timestamptz{Time: *releaseDate, Valid: true}
	} else {
		releaseDatePg = pgtype.Timestamptz{Valid: false}
	}

	config, err := ConfigServiceSingleton.queries.UpdateReleaseDate(ctx, db.UpdateReleaseDateParams{
		CoursePhaseID: coursePhaseID,
		ReleaseDate:   releaseDatePg,
		UpdatedBy:     pgtype.Text{String: updatedBy, Valid: updatedBy != ""},
	})
	if err != nil {
		log.WithError(err).Error("Failed to update release date")
		return configDTO.CoursePhaseConfig{}, err
	}

	hasDownloads, err := ConfigServiceSingleton.queries.HasDownloads(ctx, coursePhaseID)
	if err != nil {
		log.WithError(err).Warn("Failed to check for existing downloads")
		hasDownloads = false
	}

	return configDTO.MapDBConfigToDTOConfig(config, hasDownloads), nil
}

func GetTemplateContent(ctx context.Context, coursePhaseID uuid.UUID) (string, error) {
	config, err := ConfigServiceSingleton.queries.GetCoursePhaseConfig(ctx, coursePhaseID)
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
