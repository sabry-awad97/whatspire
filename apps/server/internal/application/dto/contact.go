package dto

import (
	"time"
	"whatspire/internal/domain/entity"
)

// CheckPhoneRequest represents a request to check if a phone number is on WhatsApp
type CheckPhoneRequest struct {
	SessionID string `json:"session_id" validate:"required"`
	Phone     string `json:"phone" validate:"required"`
}

// GetProfileRequest represents a request to get a user's profile
type GetProfileRequest struct {
	SessionID string `json:"session_id" validate:"required"`
	JID       string `json:"jid" validate:"required"`
}

// ContactResponse represents a contact in the response
type ContactResponse struct {
	JID          string  `json:"jid"`
	Name         string  `json:"name"`
	AvatarURL    *string `json:"avatar_url,omitempty"`
	Status       *string `json:"status,omitempty"`
	IsOnWhatsApp bool    `json:"is_on_whatsapp"`
}

// NewContactResponse creates a ContactResponse from a Contact entity
func NewContactResponse(contact *entity.Contact) *ContactResponse {
	var avatarURL *string
	if contact.AvatarURL != "" {
		avatarURL = &contact.AvatarURL
	}

	var status *string
	if contact.Status != "" {
		status = &contact.Status
	}

	return &ContactResponse{
		JID:          contact.JID,
		Name:         contact.Name,
		AvatarURL:    avatarURL,
		Status:       status,
		IsOnWhatsApp: contact.IsOnWhatsApp,
	}
}

// ContactListResponse represents a list of contacts
type ContactListResponse struct {
	Contacts []*ContactResponse `json:"contacts"`
}

// NewContactListResponse creates a ContactListResponse from a list of Contact entities
func NewContactListResponse(contacts []*entity.Contact) *ContactListResponse {
	contactResponses := make([]*ContactResponse, 0, len(contacts))
	for _, contact := range contacts {
		contactResponses = append(contactResponses, NewContactResponse(contact))
	}
	return &ContactListResponse{
		Contacts: contactResponses,
	}
}

// ChatResponse represents a chat in the response
type ChatResponse struct {
	JID             string     `json:"jid"`
	Name            string     `json:"name"`
	LastMessageTime *time.Time `json:"last_message_time,omitempty"`
	UnreadCount     int        `json:"unread_count"`
	IsGroup         bool       `json:"is_group"`
	AvatarURL       *string    `json:"avatar_url,omitempty"`
	Archived        bool       `json:"archived"`
	Pinned          bool       `json:"pinned"`
}

// NewChatResponse creates a ChatResponse from a Chat entity
func NewChatResponse(chat *entity.Chat) *ChatResponse {
	var lastMessageTime *time.Time
	if !chat.LastMessageTime.IsZero() {
		lastMessageTime = &chat.LastMessageTime
	}

	var avatarURL *string
	if chat.AvatarURL != "" {
		avatarURL = &chat.AvatarURL
	}

	return &ChatResponse{
		JID:             chat.JID,
		Name:            chat.Name,
		LastMessageTime: lastMessageTime,
		UnreadCount:     chat.UnreadCount,
		IsGroup:         chat.IsGroup,
		AvatarURL:       avatarURL,
		Archived:        chat.Archived,
		Pinned:          chat.Pinned,
	}
}

// ChatListResponse represents a list of chats
type ChatListResponse struct {
	Chats []*ChatResponse `json:"chats"`
}

// NewChatListResponse creates a ChatListResponse from a list of Chat entities
func NewChatListResponse(chats []*entity.Chat) *ChatListResponse {
	chatResponses := make([]*ChatResponse, 0, len(chats))
	for _, chat := range chats {
		chatResponses = append(chatResponses, NewChatResponse(chat))
	}
	return &ChatListResponse{
		Chats: chatResponses,
	}
}
