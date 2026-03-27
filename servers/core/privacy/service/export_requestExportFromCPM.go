package service

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	sdk "github.com/prompt-edu/prompt-sdk/promptTypes"
	log "github.com/sirupsen/logrus"
)

func RequestExportFromCPM(exportDoc ServiceExportRequest, authHeader string) ExportResult {
	requestBody := sdk.PrivacyDataExportRequest{
		PreSignedURL: exportDoc.PresignedUploadURL,
	}

	marshalledBody, err := json.Marshal(requestBody)
	if err != nil {
		log.WithError(err).Error("failed to marshal export request")
		return Failed
	}

	req, err := http.NewRequest("POST", exportDoc.APIURL, bytes.NewBuffer(marshalledBody))
	if err != nil {
		log.WithError(err).Error("failed to create export request")
		return Failed
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.WithError(err).Error("failed to call CPM export endpoint")
		return Failed
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error(err)
		}
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		return Successful
	case http.StatusNoContent:
		return SuccessfulNoData
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		log.Errorf("CPM export endpoint returned unexpected status %d: %s", resp.StatusCode, string(body))
		return Failed
	}
}
