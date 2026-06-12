package service

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	if record.UserID != userID {
		return errors.New("deletion request not found or not owned by the caller")
	}
	return nil
}

func ValidateUserMayCreateDeletionRequest(c *gin.Context) error {
	userID, err := coreutils.GetUserUUIDFromContext(c)
	if err != nil {
		return err
	}

	if _, err := PrivacyServiceSingleton.queries.GetOpenDeletionRequestForUser(c, userID); err == nil {
		return errors.New("an open deletion request already exists for this user")
	}
	return nil
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
