package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/storage/privacyexport"
)

func (s *PrivacyService) GetDownloadURLForDoc(c context.Context, docID uuid.UUID) (string, error) {
	objectKey, err := s.queries.GetExportDocObjectKey(c, docID)
	if err != nil {
		return "", err
	}

	url, err := privacyexport.GetDownloadURL(c, objectKey)
	if err != nil {
		return "", err
	}

	_ = s.queries.SetExportDocDownloadedAt(c, docID)
	return url, nil
}
