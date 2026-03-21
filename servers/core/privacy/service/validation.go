package service

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/privacy/privacyDTO"
	"github.com/prompt-edu/prompt/servers/core/utils"
)

func ValidateNoRecentExportExists(c *gin.Context, userID uuid.UUID) error {
	existing, err := GetLatestExportWithDocs(c, userID)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("a recent export already exists — please wait before requesting a new one")
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
