package service

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/privacy/privacyDTO"
	"github.com/prompt-edu/prompt/servers/core/utils"
)

func ValidateAllowedToExport(c *gin.Context, userID uuid.UUID) error {
	availability, exp, err := GetExportAvailability(c, userID)
	if err != nil { return err }
	if availability == ExportReadyForNew {
		return nil
	}
	return errors.New("Rate Limited Until " + RateLimitEndForExport(*exp).String())
}

func ValidateExportValid(c *gin.Context, exportID uuid.UUID) error {
  exp, err := PrivacyServiceSingleton.queries.GetExportRecordByID(c, exportID)
  if err != nil {
    return err
  }
  expDTO := privacyDTO.GetPrivacyExportDTOFromDBModel(exp)

  if time.Now().After(expDTO.ValidUntil) {
    return errors.New("The export is no longer valid")
  }
  return nil
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
