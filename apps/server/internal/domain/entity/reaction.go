package entity

import (
	"encoding/json"
	"time"
	"unicode/utf8"
)

// Reaction represents a WhatsApp message reaction
type Reaction struct {
	ID        string    `json:"id"`
	MessageID string    `json:"message_id"`
	SessionID string    `json:"session_id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Emoji     string    `json:"emoji"`
	Timestamp time.Time `json:"timestamp"`
}

// NewReaction creates a new Reaction
func NewReaction(id, messageID, sessionID, from, to, emoji string) *Reaction {
	return &Reaction{
		ID:        id,
		MessageID: messageID,
		SessionID: sessionID,
		From:      from,
		To:        to,
		Emoji:     emoji,
		Timestamp: time.Now(),
	}
}

// IsValid checks if the reaction is valid
func (r *Reaction) IsValid() bool {
	if r.ID == "" || r.MessageID == "" || r.SessionID == "" {
		return false
	}
	if r.From == "" || r.To == "" {
		return false
	}
	return r.IsValidEmoji()
}

// IsValidEmoji checks if the emoji is a valid Unicode emoji
func (r *Reaction) IsValidEmoji() bool {
	// Empty emoji is valid (used for removing reactions)
	if r.Emoji == "" {
		return true
	}

	// Check if it's valid UTF-8
	if !utf8.ValidString(r.Emoji) {
		return false
	}

	// Check if it's a single emoji (1-4 runes for emoji with modifiers)
	runeCount := utf8.RuneCountInString(r.Emoji)
	if runeCount < 1 || runeCount > 4 {
		return false
	}

	// Basic check: emoji should contain at least one emoji-range character
	// This is a simplified check - full emoji validation is complex
	for _, r := range r.Emoji {
		// Check common emoji ranges
		if (r >= 0x1F600 && r <= 0x1F64F) || // Emoticons
			(r >= 0x1F300 && r <= 0x1F5FF) || // Misc Symbols and Pictographs
			(r >= 0x1F680 && r <= 0x1F6FF) || // Transport and Map
			(r >= 0x2600 && r <= 0x26FF) || // Misc symbols
			(r >= 0x2700 && r <= 0x27BF) || // Dingbats
			(r >= 0xFE00 && r <= 0xFE0F) || // Variation Selectors
			(r >= 0x1F900 && r <= 0x1F9FF) || // Supplemental Symbols and Pictographs
			(r >= 0x1FA70 && r <= 0x1FAFF) || // Symbols and Pictographs Extended-A
			(r >= 0x200D) { // Zero Width Joiner (for compound emoji)
			return true
		}
	}

	return false
}

// IsRemoval returns true if this reaction is a removal (empty emoji)
func (r *Reaction) IsRemoval() bool {
	return r.Emoji == ""
}

// MarshalJSON implements json.Marshaler
func (r *Reaction) MarshalJSON() ([]byte, error) {
	type Alias Reaction
	return json.Marshal(&struct {
		*Alias
		Timestamp string `json:"timestamp"`
	}{
		Alias:     (*Alias)(r),
		Timestamp: r.Timestamp.Format(time.RFC3339),
	})
}
