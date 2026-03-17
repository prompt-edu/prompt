package developerProfile

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/intro_course/db/sqlc"
	"github.com/prompt-edu/prompt/servers/intro_course/developerProfile/developerProfileDTO"
	log "github.com/sirupsen/logrus"
)

type DeveloperProfileService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var DeveloperProfileServiceSingleton *DeveloperProfileService

func CreateDeveloperProfile(ctx context.Context, coursePhaseID uuid.UUID, courseParticipationID uuid.UUID, request developerProfileDTO.PostDeveloperProfile) error {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	params := developerProfileDTO.GetDeveloperProfileDTOFromPostRequest(request, coursePhaseID, courseParticipationID)
	err := DeveloperProfileServiceSingleton.queries.CreateDeveloperProfile(ctxWithTimeout, params)
	if err != nil {
		log.WithFields(log.Fields{
			"coursePhaseID":         coursePhaseID,
			"courseParticipationID": courseParticipationID,
			"error":                 err,
		}).Error("Failed to create developer profile")
		return errors.New("failed to create developer profile")
	}
	return nil

}

func GetOwnDeveloperProfile(ctx context.Context, coursePhaseID uuid.UUID, courseParticipationID uuid.UUID) (developerProfileDTO.DeveloperProfile, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	params := db.GetDeveloperProfileByCourseParticipationIDParams{
		CoursePhaseID:         coursePhaseID,
		CourseParticipationID: courseParticipationID,
	}
	developerProfile, err := DeveloperProfileServiceSingleton.queries.GetDeveloperProfileByCourseParticipationID(ctxWithTimeout, params)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		// return empty developer profile if no profile exists
		return developerProfileDTO.DeveloperProfile{}, nil
	} else if err != nil {
		log.WithFields(log.Fields{
			"coursePhaseID":         coursePhaseID,
			"courseParticipationID": courseParticipationID,
			"error":                 err,
		}).Error("Failed to get developer profile")
		return developerProfileDTO.DeveloperProfile{}, errors.New("failed to get developer profile")
	}

	return developerProfileDTO.GetDeveloperProfileDTOFromDBModel(developerProfile), nil
}

func GetAllDeveloperProfiles(ctx context.Context, coursePhaseID uuid.UUID) ([]developerProfileDTO.DeveloperProfile, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	developerProfiles, err := DeveloperProfileServiceSingleton.queries.GetAllDeveloperProfiles(ctxWithTimeout, coursePhaseID)
	if err != nil {
		log.WithFields(log.Fields{
			"coursePhaseID": coursePhaseID,
			"error":         err,
		}).Error("Failed to get developer profiles")
		return nil, err
	}

	developerProfileDTOs := developerProfileDTO.GetDeveloperProfileDTOsFromDBModels(developerProfiles)
	return developerProfileDTOs, nil
}

func CreateOrUpdateDeveloperProfile(ctx context.Context, coursePhaseID uuid.UUID, courseParticipationID uuid.UUID, request developerProfileDTO.DeveloperProfile) error {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	params := developerProfileDTO.GetDeveloperProfileDTOFromCreateRequest(request, coursePhaseID, courseParticipationID)
	err := DeveloperProfileServiceSingleton.queries.CreateOrUpdateDeveloperProfile(ctxWithTimeout, params)
	if err != nil {
		log.WithFields(log.Fields{
			"coursePhaseID":         coursePhaseID,
			"courseParticipationID": courseParticipationID,
			"error":                 err,
		}).Error("Failed to create or update developer profile")
		return errors.New("failed to create or update developer profile")
	}
	return nil
}

func GetDevicesForCoursePhase(ctx context.Context, coursePhaseID uuid.UUID) ([]developerProfileDTO.DeviceWithParticipationID, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	devicesWithParticipations, err := DeveloperProfileServiceSingleton.queries.GetDevicesForCoursePhase(ctxWithTimeout, coursePhaseID)
	if err != nil {
		log.WithFields(log.Fields{
			"coursePhaseID": coursePhaseID,
			"error":         err,
		}).Error("Failed to get devices with participation ID")
		return nil, errors.New("failed to get devices with participation ID")
	}

	return developerProfileDTO.GetDeviceWithParticipationIDFromDBModel(devicesWithParticipations), nil
}

func GetDevicesForCourseParticipation(ctx context.Context, coursePhaseID uuid.UUID, courseParticipationID uuid.UUID) ([]string, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	params := db.GetDevicesForCourseParticipationParams{
		CoursePhaseID:         coursePhaseID,
		CourseParticipationID: courseParticipationID,
	}
	devices, err := DeveloperProfileServiceSingleton.queries.GetDevicesForCourseParticipation(ctxWithTimeout, params)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, err // return the error to write a 404
	} else if err != nil {
		log.WithFields(log.Fields{
			"coursePhaseID":         coursePhaseID,
			"courseParticipationID": courseParticipationID,
			"error":                 err,
		}).Error("Failed to get devices with participation ID")
		return nil, errors.New("failed to get devices with participation ID")
	}

	return devices, nil
}
