package presentation

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type eventSubscription struct {
	id     uuid.UUID
	userID string
	name   string
	ch     chan FeedbackEvent
}

type EventHub struct {
	mu            sync.RWMutex
	subscriptions map[uuid.UUID]map[uuid.UUID]*eventSubscription
}

func NewEventHub() *EventHub {
	return &EventHub{subscriptions: make(map[uuid.UUID]map[uuid.UUID]*eventSubscription)}
}

func (h *EventHub) Subscribe(presentationID uuid.UUID, userID, name string) (uuid.UUID, <-chan FeedbackEvent, func()) {
	h.mu.Lock()
	connectionID := uuid.New()
	subscription := &eventSubscription{
		id:     connectionID,
		userID: userID,
		name:   name,
		ch:     make(chan FeedbackEvent, 16),
	}
	if h.subscriptions[presentationID] == nil {
		h.subscriptions[presentationID] = make(map[uuid.UUID]*eventSubscription)
	}
	h.subscriptions[presentationID][connectionID] = subscription
	h.mu.Unlock()

	h.publishPresence(presentationID)
	return connectionID, subscription.ch, func() {
		h.mu.Lock()
		if subscriptions := h.subscriptions[presentationID]; subscriptions != nil {
			delete(subscriptions, connectionID)
			if len(subscriptions) == 0 {
				delete(h.subscriptions, presentationID)
			}
		}
		h.mu.Unlock()
		h.publishPresence(presentationID)
	}
}

func (h *EventHub) ActiveEditors(presentationID uuid.UUID) []ActiveEditorResponse {
	h.mu.RLock()
	defer h.mu.RUnlock()

	subscriptions := h.subscriptions[presentationID]
	editors := make([]ActiveEditorResponse, 0, len(subscriptions))
	expiresAt := time.Now().Add(45 * time.Second)
	for _, subscription := range subscriptions {
		editors = append(editors, ActiveEditorResponse{
			ConnectionID: subscription.id,
			UserID:       subscription.userID,
			Name:         subscription.name,
			ExpiresAt:    expiresAt,
		})
	}
	return editors
}

func (h *EventHub) Publish(event FeedbackEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, subscription := range h.subscriptions[event.PresentationID] {
		select {
		case subscription.ch <- event:
		default:
		}
	}
}

func (h *EventHub) publishPresence(presentationID uuid.UUID) {
	h.Publish(FeedbackEvent{
		Type:           "presence.updated",
		PresentationID: presentationID,
		ActiveEditors:  h.ActiveEditors(presentationID),
	})
}
