package presentation

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/presentation/db/sqlc"
)

func TestValidateSlot(t *testing.T) {
	start := time.Date(2026, time.July, 16, 10, 0, 0, 0, time.UTC)
	if err := validateSlot(SlotRequest{StartTime: start, EndTime: start.Add(15 * time.Minute)}); err != nil {
		t.Fatalf("expected valid slot, got %v", err)
	}
	if err := validateSlot(SlotRequest{StartTime: start, EndTime: start}); err == nil {
		t.Fatal("expected equal start and end to be rejected")
	}
	if err := validateSlot(SlotRequest{StartTime: start, EndTime: start.Add(-time.Minute)}); err == nil {
		t.Fatal("expected end before start to be rejected")
	}
}

func TestMaterialMutationDeadline(t *testing.T) {
	now := time.Date(2026, time.July, 16, 10, 0, 0, 0, time.UTC)
	service := &Service{now: func() time.Time { return now }}
	futureSlot := db.PresentationSlot{
		StartTime: pgtype.Timestamptz{Time: now.Add(time.Minute), Valid: true},
	}
	startedSlot := db.PresentationSlot{
		StartTime: pgtype.Timestamptz{Time: now, Valid: true},
	}

	if err := service.ensureMaterialMutationAllowed(User{}, futureSlot); err != nil {
		t.Fatalf("expected student upload before presentation start to be allowed, got %v", err)
	}
	if err := service.ensureMaterialMutationAllowed(User{}, startedSlot); err == nil {
		t.Fatal("expected student upload at presentation start to be rejected")
	}
	if err := service.ensureMaterialMutationAllowed(User{Staff: true}, startedSlot); err != nil {
		t.Fatalf("expected staff upload after presentation start to be allowed, got %v", err)
	}
}

func TestFeedbackScope(t *testing.T) {
	service := &Service{}
	user := User{ID: "instructor-1", Name: "Ada Lovelace"}

	scope, evaluatorID, evaluatorName := service.feedbackScope(
		SettingsResponse{FeedbackMode: feedbackIndependent},
		user,
	)
	if scope != user.ID || !evaluatorID.Valid || evaluatorID.String != user.ID || evaluatorName != user.Name {
		t.Fatalf("unexpected independent feedback scope: %q %#v %q", scope, evaluatorID, evaluatorName)
	}

	scope, evaluatorID, evaluatorName = service.feedbackScope(
		SettingsResponse{FeedbackMode: feedbackShared},
		user,
	)
	if scope != "shared" || evaluatorID.Valid || evaluatorName != "Shared feedback" {
		t.Fatalf("unexpected shared feedback scope: %q %#v %q", scope, evaluatorID, evaluatorName)
	}
}

func TestPresentationResponseIncludesNamedRelease(t *testing.T) {
	releaseName := "Final jury feedback"
	releasedAt := time.Date(2026, time.July, 16, 12, 0, 0, 0, time.UTC)
	presentation := db.Presentation{
		ID:                  uuid.New(),
		FeedbackReleaseName: pgtype.Text{String: releaseName, Valid: true},
		FeedbackReleasedAt:  pgtype.Timestamptz{Time: releasedAt, Valid: true},
	}
	if !presentation.FeedbackReleaseName.Valid || presentation.FeedbackReleaseName.String != releaseName {
		t.Fatal("expected named release metadata to be persisted in the presentation model")
	}
}
