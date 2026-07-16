package presentation

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEventHubPublishesPresenceAndAnswers(t *testing.T) {
	hub := NewEventHub()
	presentationID := uuid.New()
	connectionID, events, unsubscribe := hub.Subscribe(presentationID, "user-1", "Grace Hopper")
	defer unsubscribe()

	select {
	case event := <-events:
		if event.Type != "presence.updated" || len(event.ActiveEditors) != 1 {
			t.Fatalf("unexpected presence event: %#v", event)
		}
		if event.ActiveEditors[0].ConnectionID != connectionID {
			t.Fatalf("unexpected connection ID: %s", event.ActiveEditors[0].ConnectionID)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for presence event")
	}

	answer := FeedbackAnswerResponse{CategoryID: uuid.New(), Value: "Clear structure", Revision: 1}
	hub.Publish(FeedbackEvent{
		Type:           "answer.updated",
		PresentationID: presentationID,
		Answer:         &answer,
	})
	select {
	case event := <-events:
		if event.Type != "answer.updated" || event.Answer == nil || event.Answer.Revision != 1 {
			t.Fatalf("unexpected answer event: %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for answer event")
	}
}

func TestEventHubTracksIndependentConnections(t *testing.T) {
	hub := NewEventHub()
	presentationID := uuid.New()
	_, firstEvents, unsubscribeFirst := hub.Subscribe(presentationID, "user-1", "First")
	defer unsubscribeFirst()
	<-firstEvents
	_, secondEvents, unsubscribeSecond := hub.Subscribe(presentationID, "user-2", "Second")
	defer unsubscribeSecond()
	<-secondEvents

	if editors := hub.ActiveEditors(presentationID); len(editors) != 2 {
		t.Fatalf("expected two active editors, got %d", len(editors))
	}
}
