package dto

import (
	"fmt"
)

// EventDTO represents an event in API responses
type EventDTO struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	SessionID string `json:"session_id"`
	Data      []byte `json:"data,omitempty"`
	Timestamp string `json:"timestamp"`
}

// QueryEventsRequest represents a request to query events
type QueryEventsRequest struct {
	SessionID string `json:"session_id,omitempty" form:"session_id"`
	EventType string `json:"event_type,omitempty" form:"event_type"`
	Since     string `json:"since,omitempty" form:"since"` // RFC3339 timestamp
	Until     string `json:"until,omitempty" form:"until"` // RFC3339 timestamp
	Limit     int    `json:"limit,omitempty" form:"limit"`
	Offset    int    `json:"offset,omitempty" form:"offset"`
}

// Validate validates the query events request
func (r *QueryEventsRequest) Validate() error {
	// Set default limit if not specified
	if r.Limit == 0 {
		r.Limit = 100
	}

	// Validate limit
	if r.Limit < 0 || r.Limit > 1000 {
		return fmt.Errorf("limit must be between 1 and 1000")
	}

	// Validate offset
	if r.Offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}

	return nil
}

// QueryEventsResponse represents a response to query events
type QueryEventsResponse struct {
	Events []EventDTO `json:"events"`
	Total  int64      `json:"total"`
	Limit  int        `json:"limit"`
	Offset int        `json:"offset"`
}

// ReplayEventsRequest represents a request to replay events
type ReplayEventsRequest struct {
	SessionID string `json:"session_id,omitempty"`
	EventType string `json:"event_type,omitempty"`
	Since     string `json:"since,omitempty"` // RFC3339 timestamp
	Until     string `json:"until,omitempty"` // RFC3339 timestamp
	DryRun    bool   `json:"dry_run,omitempty"`
}

// Validate validates the replay events request
func (r *ReplayEventsRequest) Validate() error {
	// At least one filter must be specified to prevent replaying all events
	if r.SessionID == "" && r.EventType == "" && r.Since == "" && r.Until == "" {
		return fmt.Errorf("at least one filter (session_id, event_type, since, or until) must be specified")
	}

	return nil
}

// ReplayEventsResponse represents a response to replay events
type ReplayEventsResponse struct {
	Success        bool   `json:"success"`
	EventsFound    int    `json:"events_found"`
	EventsReplayed int    `json:"events_replayed"`
	EventsFailed   int    `json:"events_failed,omitempty"`
	DryRun         bool   `json:"dry_run"`
	Message        string `json:"message"`
	LastError      error  `json:"last_error,omitempty"`
}
