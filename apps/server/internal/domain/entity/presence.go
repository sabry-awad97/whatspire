package entity

import (
	"encoding/json"
	"time"
)

// PresenceState represents the presence state
type PresenceState string

const (
	PresenceStateTyping  PresenceState = "typing"
	PresenceStatePaused  PresenceState = "paused"
	PresenceStateOnline  PresenceState = "online"
	PresenceStateOffline PresenceState = "offline"
)

// IsValid checks if the presence state is valid
func (ps PresenceState) IsValid() bool {
	switch ps {
	case PresenceStateTyping, PresenceStatePaused, PresenceStateOnline, PresenceStateOffline:
		return true
	}
	return false
}

// String returns the string representation of the presence state
func (ps PresenceState) String() string {
	return string(ps)
}

// Presence represents a WhatsApp presence update
type Presence struct {
	ID        string        `json:"id"`
	SessionID string        `json:"session_id"`
	UserJID   string        `json:"user_jid"` // The user whose presence changed
	ChatJID   string        `json:"chat_jid"` // The chat where presence is shown (can be empty for general presence)
	State     PresenceState `json:"state"`    // typing, paused, online, offline
	Timestamp time.Time     `json:"timestamp"`
}

// NewPresence creates a new Presence
func NewPresence(id, sessionID, userJID, chatJID string, state PresenceState) *Presence {
	return &Presence{
		ID:        id,
		SessionID: sessionID,
		UserJID:   userJID,
		ChatJID:   chatJID,
		State:     state,
		Timestamp: time.Now(),
	}
}

// IsValid checks if the presence is valid
func (p *Presence) IsValid() bool {
	if p.ID == "" || p.SessionID == "" || p.UserJID == "" {
		return false
	}
	return p.State.IsValid()
}

// IsTyping returns true if the presence state is typing
func (p *Presence) IsTyping() bool {
	return p.State == PresenceStateTyping
}

// IsPaused returns true if the presence state is paused
func (p *Presence) IsPaused() bool {
	return p.State == PresenceStatePaused
}

// IsOnline returns true if the presence state is online
func (p *Presence) IsOnline() bool {
	return p.State == PresenceStateOnline
}

// IsOffline returns true if the presence state is offline
func (p *Presence) IsOffline() bool {
	return p.State == PresenceStateOffline
}

// MarshalJSON implements json.Marshaler
func (p *Presence) MarshalJSON() ([]byte, error) {
	type Alias Presence
	return json.Marshal(&struct {
		*Alias
		Timestamp string `json:"timestamp"`
	}{
		Alias:     (*Alias)(p),
		Timestamp: p.Timestamp.Format(time.RFC3339),
	})
}
