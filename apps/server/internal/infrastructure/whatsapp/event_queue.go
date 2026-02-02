package whatsapp

import (
	"sync"

	"whatspire/internal/domain/entity"
)

// EventQueue manages queued events for disconnected sessions
// Events are queued when a session is disconnected and flushed when it reconnects
type EventQueue struct {
	mu     sync.RWMutex
	queues map[string][]*entity.Event // sessionID -> events
}

// NewEventQueue creates a new event queue
func NewEventQueue() *EventQueue {
	return &EventQueue{
		queues: make(map[string][]*entity.Event),
	}
}

// Enqueue adds an event to the queue for its session
func (eq *EventQueue) Enqueue(event *entity.Event) {
	eq.mu.Lock()
	defer eq.mu.Unlock()

	if eq.queues[event.SessionID] == nil {
		eq.queues[event.SessionID] = make([]*entity.Event, 0)
	}

	eq.queues[event.SessionID] = append(eq.queues[event.SessionID], event)
}

// GetSessionEvents returns all queued events for a session
func (eq *EventQueue) GetSessionEvents(sessionID string) []*entity.Event {
	eq.mu.RLock()
	defer eq.mu.RUnlock()

	events := eq.queues[sessionID]
	if events == nil {
		return []*entity.Event{}
	}

	// Return a copy to prevent external modification
	result := make([]*entity.Event, len(events))
	copy(result, events)
	return result
}

// FlushSession removes all queued events for a session and returns them
func (eq *EventQueue) FlushSession(sessionID string) []*entity.Event {
	eq.mu.Lock()
	defer eq.mu.Unlock()

	events := eq.queues[sessionID]
	delete(eq.queues, sessionID)

	if events == nil {
		return []*entity.Event{}
	}

	return events
}

// Clear removes all queued events for a session without returning them
func (eq *EventQueue) Clear(sessionID string) {
	eq.mu.Lock()
	defer eq.mu.Unlock()

	delete(eq.queues, sessionID)
}

// Size returns the number of queued events for a session
func (eq *EventQueue) Size(sessionID string) int {
	eq.mu.RLock()
	defer eq.mu.RUnlock()

	return len(eq.queues[sessionID])
}

// TotalSize returns the total number of queued events across all sessions
func (eq *EventQueue) TotalSize() int {
	eq.mu.RLock()
	defer eq.mu.RUnlock()

	total := 0
	for _, events := range eq.queues {
		total += len(events)
	}
	return total
}

// Sessions returns all session IDs that have queued events
func (eq *EventQueue) Sessions() []string {
	eq.mu.RLock()
	defer eq.mu.RUnlock()

	sessions := make([]string, 0, len(eq.queues))
	for sessionID := range eq.queues {
		sessions = append(sessions, sessionID)
	}
	return sessions
}
