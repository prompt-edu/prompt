package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/storage/privacyexport"
)

func GetDownloadURLForDoc(c *gin.Context, exportID uuid.UUID, docID uuid.UUID) (string, error) {
	exp, err := GetExportWithDocs(c, exportID)
	if err != nil {
		return "", err
	}

	for _, doc := range exp.Documents {
		if doc.ID == docID {
			return privacyexport.GetDownloadURL(c.Request.Context(), doc.ObjectKey)
		}
	}

	return "", fmt.Errorf("document %s not found in export %s", docID, exportID)
}
