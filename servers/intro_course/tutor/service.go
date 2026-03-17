package tutor

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/intro_course/db/sqlc"
	"github.com/prompt-edu/prompt/servers/intro_course/tutor/tutorDTO"
	log "github.com/sirupsen/logrus"
)

type TutorService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var TutorServiceSingleton *TutorService

func GetTutors(ctx context.Context, coursePhaseID uuid.UUID) ([]tutorDTO.Tutor, error) {
	tutors, err := TutorServiceSingleton.queries.GetAllTutors(ctx, coursePhaseID)
	if err != nil {
		log.Error("Error getting tutors: ", err)
		return nil, errors.New("error getting tutors")
	}

	return tutorDTO.GetTutorDTOsFromModels(tutors), nil
}

func ImportTutors(ctx context.Context, coursePhaseID uuid.UUID, tutors []tutorDTO.Tutor) error {
	// add students to the keycloak group
	tx, err := TutorServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer sdkUtils.DeferRollback(tx, ctx)
	qtx := TutorServiceSingleton.queries.WithTx(tx)

	for _, tutor := range tutors {
		// store tutor in database
		err := qtx.CreateTutor(ctx, db.CreateTutorParams{
			CoursePhaseID:       coursePhaseID,
			ID:                  tutor.ID,
			FirstName:           tutor.FirstName,
			LastName:            tutor.LastName,
			Email:               tutor.Email,
			MatriculationNumber: tutor.MatriculationNumber,
			UniversityLogin:     tutor.UniversityLogin,
		})
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func UpdateGitLabUsername(ctx context.Context, coursePhaseID, tutorID uuid.UUID, updateTutor tutorDTO.UpdateTutor) error {
	err := TutorServiceSingleton.queries.UpdateTutorGitlabUsername(ctx, db.UpdateTutorGitlabUsernameParams{
		CoursePhaseID:  coursePhaseID,
		ID:             tutorID,
		GitlabUsername: pgtype.Text{String: updateTutor.GitlabUsername, Valid: true},
	})
	if err != nil {
		log.Error("Error updating tutor: ", err)
		return errors.New("error updating tutor")
	}

	return nil
}
