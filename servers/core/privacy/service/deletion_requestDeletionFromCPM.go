package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	sdk "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	sdkTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
	log "github.com/sirupsen/logrus"
)

const cpmDeletionRequestTimeout = 5 * time.Minute

func RequestDeletionFromCPM(ctx context.Context, apiURL string, subject sdk.SubjectIdentifiers, authHeader string) error {
	body := sdkTypes.PrivacyDeletionRequest{SubjectIdentifiers: subject}
	marshalled, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal deletion request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(marshalled))
	if err != nil {
		return fmt.Errorf("failed to create deletion request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	client := &http.Client{Timeout: cpmDeletionRequestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call CPM deletion endpoint: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error(err)
		}
	}()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	return fmt.Errorf("CPM deletion endpoint returned status %d: %s", resp.StatusCode, string(respBody))
}
