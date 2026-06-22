package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/storage/privacyexport"
)

func GetDownloadURLForDoc(c context.Context, docID uuid.UUID) (string, error) {
	objectKey, err := PrivacyServiceSingleton.queries.GetExportDocObjectKey(c, docID)
	if err != nil {
		return "", err
	}

	url, err := privacyexport.GetDownloadURL(c, objectKey)
	if err != nil {
		return "", err
	}

	_ = PrivacyServiceSingleton.queries.SetExportDocDownloadedAt(c, docID)
	return url, nil
}
