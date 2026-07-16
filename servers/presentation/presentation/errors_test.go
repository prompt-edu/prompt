package presentation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestFeedbackConflictResponseExposesCurrentAnswer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	answer := FeedbackAnswerResponse{CategoryID: uuid.New(), Value: "New value", Revision: 4}

	writeError(context, &APIError{
		Status:  http.StatusConflict,
		Code:    "feedback_conflict",
		Message: "Feedback was updated by another editor",
		Details: answer,
	})

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", recorder.Code)
	}
	var body struct {
		Code   string                 `json:"code"`
		Answer FeedbackAnswerResponse `json:"answer"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode conflict response: %v", err)
	}
	if body.Code != "feedback_conflict" || body.Answer.Revision != answer.Revision {
		t.Fatalf("unexpected conflict response: %#v", body)
	}
}
