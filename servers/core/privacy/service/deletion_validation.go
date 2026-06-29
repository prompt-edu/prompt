package service

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	coreutils "github.com/prompt-edu/prompt/servers/core/utils"
)

func ValidateDeletionRequestBelongsToCaller(c *gin.Context, requestID uuid.UUID) error {
	userID, err := coreutils.GetUserUUIDFromContext(c)
	if err != nil {
		return err
	}

	record, err := PrivacyServiceSingleton.queries.GetDeletionRequestByID(c, requestID)
	if err != nil {
		return errors.New("deletion request not found or not owned by the caller")
	}
	if !record.UserID.Valid || record.UserID.Bytes != userID {
		return errors.New("deletion request not found or not owned by the caller")
	}
	return nil
}

func ValidateUserMayCreateDeletionRequest(c *gin.Context) error {
	userID, err := coreutils.GetUserUUIDFromContext(c)
	if err != nil {
		return err
	}

	_, err = PrivacyServiceSingleton.queries.GetOpenDeletionRequestForUser(c, pgtype.UUID{Bytes: userID, Valid: true})
	if err == nil {
		return errors.New("an open deletion request already exists for this user")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	return nil
}

func ValidateStudentsExist(c *gin.Context, studentIDs []uuid.UUID) error {
	existing, err := PrivacyServiceSingleton.queries.GetExistingStudentIDs(c, studentIDs)
	if err != nil {
		return err
	}
	if len(existing) == len(studentIDs) {
		return nil
	}

	existingSet := make(map[uuid.UUID]struct{}, len(existing))
	for _, id := range existing {
		existingSet[id] = struct{}{}
	}
	missing := make([]string, 0, len(studentIDs)-len(existing))
	for _, id := range studentIDs {
		if _, ok := existingSet[id]; !ok {
			missing = append(missing, id.String())
		}
	}
	return fmt.Errorf("unknown student_ids: %v", missing)
}

func ValidateDeletionRequestPending(c *gin.Context, requestID uuid.UUID) error {
	record, err := PrivacyServiceSingleton.queries.GetDeletionRequestByID(c, requestID)
	if err != nil {
		return errors.New("deletion request not found or no longer in pending_approval state")
	}
	if record.Status != db.PrivacyDeletionRequestStatusPendingApproval {
		return errors.New("deletion request not found or no longer in pending_approval state")
	}
	return nil
}
