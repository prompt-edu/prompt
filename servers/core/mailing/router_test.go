package mailing

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/mailing/mailingDTO"
	"github.com/stretchr/testify/assert"
)

func setupMailingTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	setupMailingRouter(api, func() gin.HandlerFunc {
		return testutils.MockAuthMiddleware([]string{"PROMPT_Admin"})
	}, testutils.MockPermissionMiddleware)

	return router
}

func TestSendManualMailTriggerInvalidPhaseID(t *testing.T) {
	router := setupMailingTestRouter()
	requestBody := []byte(`{"subject":"s","content":"c","recipientCourseParticipationIDs":[]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/mailing/not-a-uuid/manual", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, resp.Body.String(), "invalid UUID")
}

func TestSendManualMailTriggerInvalidBody(t *testing.T) {
	router := setupMailingTestRouter()
	phaseID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/mailing/"+phaseID.String()+"/manual",
		bytes.NewReader([]byte(`{`)),
	)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestSendManualMailTriggerSuccess(t *testing.T) {
	oldSendManualMailFn := sendManualMailFn
	t.Cleanup(func() { sendManualMailFn = oldSendManualMailFn })

	sendManualMailFn = func(
		ctx context.Context,
		coursePhaseID uuid.UUID,
		request mailingDTO.SendManualMailRequest,
	) (mailingDTO.ManualMailReport, error) {
		assert.Equal(t, "Subj", request.Subject)
		assert.Equal(t, 1, len(request.RecipientCourseParticipationIDs))
		return mailingDTO.ManualMailReport{
			SuccessfulEmails:    []string{"alice@example.com"},
			FailedEmails:        []string{},
			RequestedRecipients: 1,
			SentAt:              time.Date(2026, time.January, 1, 10, 0, 0, 0, time.UTC),
		}, nil
	}

	router := setupMailingTestRouter()
	phaseID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	requestBody := []byte(`{
		"subject":"Subj",
		"content":"Body",
		"recipientCourseParticipationIDs":["44444444-4444-4444-4444-444444444444"],
		"additionalPlaceholders":{"k":"v"}
	}`)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/mailing/"+phaseID.String()+"/manual",
		bytes.NewReader(requestBody),
	)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var report mailingDTO.ManualMailReport
	err := json.Unmarshal(resp.Body.Bytes(), &report)
	assert.NoError(t, err)
	assert.Equal(t, 1, report.RequestedRecipients)
}

func TestSendManualMailTriggerServiceError(t *testing.T) {
	oldSendManualMailFn := sendManualMailFn
	t.Cleanup(func() { sendManualMailFn = oldSendManualMailFn })

	sendManualMailFn = func(
		ctx context.Context,
		coursePhaseID uuid.UUID,
		request mailingDTO.SendManualMailRequest,
	) (mailingDTO.ManualMailReport, error) {
		return mailingDTO.ManualMailReport{}, errors.New("sending failed")
	}

	router := setupMailingTestRouter()
	phaseID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	requestBody := []byte(`{"subject":"Subj","content":"Body","recipientCourseParticipationIDs":[]}`)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/mailing/"+phaseID.String()+"/manual",
		bytes.NewReader(requestBody),
	)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Contains(t, resp.Body.String(), "sending failed")
}

func TestSendManualMailTriggerValidationError(t *testing.T) {
	oldSendManualMailFn := sendManualMailFn
	t.Cleanup(func() { sendManualMailFn = oldSendManualMailFn })

	sendManualMailFn = func(
		ctx context.Context,
		coursePhaseID uuid.UUID,
		request mailingDTO.SendManualMailRequest,
	) (mailingDTO.ManualMailReport, error) {
		return mailingDTO.ManualMailReport{}, ErrManualMailValidation
	}

	router := setupMailingTestRouter()
	phaseID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	requestBody := []byte(`{"subject":"Subj","content":"Body","recipientCourseParticipationIDs":[]}`)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/mailing/"+phaseID.String()+"/manual",
		bytes.NewReader(requestBody),
	)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, resp.Body.String(), ErrManualMailValidation.Error())
}
