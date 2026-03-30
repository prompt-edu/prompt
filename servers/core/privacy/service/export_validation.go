package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/privacy/privacyDTO"
	"github.com/prompt-edu/prompt/servers/core/utils"
)

func ValidateExportValid(c context.Context, exportID uuid.UUID) error {
	exp, err := PrivacyServiceSingleton.queries.GetExportRecordByID(c, exportID)
	if err != nil {
		return err
	}
	expDTO := privacyDTO.GetPrivacyExportDTOFromDBModel(exp)

	if time.Now().After(expDTO.ValidUntil) {
		return errors.New("the export is no longer valid")
	}
	return nil
}

func ValidateExportDocBelongsToExport(c context.Context, docID uuid.UUID, exportID uuid.UUID) error {
	expWD, err := GetExportWithDocs(c, exportID)
	if err != nil {
		return err
	}
	for _, expDoc := range expWD.Documents {
		if expDoc.ID == docID {
			return nil
		}
	}
	return fmt.Errorf("the export doc with given ID does not belong to the export with given ID")
}

func ValidateExportBelongsToRequester(c *gin.Context, exportID uuid.UUID) error {
	exp, err := PrivacyServiceSingleton.queries.GetExportRecordByID(c, exportID)
	if err != nil {
		return err
	}

	expDTO := privacyDTO.GetPrivacyExportDTOFromDBModel(exp)

	requesterUserID, errGetID := utils.GetUserUUIDFromContext(c)
	if errGetID != nil {
		return errGetID
	}

	if expDTO.UserID != requesterUserID {
		return errors.New("export does not belong to the requester")
	}

	return nil
}
