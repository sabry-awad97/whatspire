package entity

import (
	"encoding/json"
	"time"
)

// Contact represents a WhatsApp contact
type Contact struct {
	JID          string `json:"jid"`
	Name         string `json:"name"`
	AvatarURL    string `json:"avatar_url,omitempty"`
	Status       string `json:"status,omitempty"`
	IsOnWhatsApp bool   `json:"is_on_whatsapp"`
}

// NewContact creates a new Contact
func NewContact(jid, name string, isOnWhatsApp bool) *Contact {
	return &Contact{
		JID:          jid,
		Name:         name,
		IsOnWhatsApp: isOnWhatsApp,
	}
}

// IsValid checks if the contact is valid
func (c *Contact) IsValid() bool {
	return c.JID != ""
}

// SetAvatar sets the avatar URL for the contact
func (c *Contact) SetAvatar(url string) {
	c.AvatarURL = url
}

// SetStatus sets the status message for the contact
func (c *Contact) SetStatus(status string) {
	c.Status = status
}

// Chat represents a WhatsApp chat (individual or group)
type Chat struct {
	JID             string    `json:"jid"`
	Name            string    `json:"name"`
	LastMessageTime time.Time `json:"last_message_time"`
	UnreadCount     int       `json:"unread_count"`
	IsGroup         bool      `json:"is_group"`
	AvatarURL       string    `json:"avatar_url,omitempty"`
	Archived        bool      `json:"archived"`
	Pinned          bool      `json:"pinned"`
}

// NewChat creates a new Chat
func NewChat(jid, name string, isGroup bool) *Chat {
	return &Chat{
		JID:             jid,
		Name:            name,
		IsGroup:         isGroup,
		LastMessageTime: time.Time{},
		UnreadCount:     0,
		Archived:        false,
		Pinned:          false,
	}
}

// IsValid checks if the chat is valid
func (c *Chat) IsValid() bool {
	return c.JID != ""
}

// SetLastMessageTime sets the last message time for the chat
func (c *Chat) SetLastMessageTime(t time.Time) {
	c.LastMessageTime = t
}

// SetUnreadCount sets the unread count for the chat
func (c *Chat) SetUnreadCount(count int) {
	if count < 0 {
		count = 0
	}
	c.UnreadCount = count
}

// SetAvatar sets the avatar URL for the chat
func (c *Chat) SetAvatar(url string) {
	c.AvatarURL = url
}

// SetArchived sets the archived status for the chat
func (c *Chat) SetArchived(archived bool) {
	c.Archived = archived
}

// SetPinned sets the pinned status for the chat
func (c *Chat) SetPinned(pinned bool) {
	c.Pinned = pinned
}

// MarshalJSON implements json.Marshaler for Chat
func (c *Chat) MarshalJSON() ([]byte, error) {
	type Alias Chat
	return json.Marshal(&struct {
		*Alias
		LastMessageTime string `json:"last_message_time"`
	}{
		Alias:           (*Alias)(c),
		LastMessageTime: c.LastMessageTime.Format(time.RFC3339),
	})
}
