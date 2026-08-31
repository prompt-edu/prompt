package service

import (
	"context"

	"github.com/google/uuid"
)

func (s *PrivacyService) GetDownloadURLForDoc(c context.Context, docID uuid.UUID) (string, error) {
	objectKey, err := s.queries.GetExportDocObjectKey(c, docID)
	if err != nil {
		return "", err
	}

	url, err := s.exportStorage.GetDownloadURL(c, objectKey)
	if err != nil {
		return "", err
	}

	_ = s.queries.SetExportDocDownloadedAt(c, docID)
	return url, nil
}
