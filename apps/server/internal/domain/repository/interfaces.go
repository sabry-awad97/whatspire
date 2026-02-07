package repository

import (
	"context"

	"whatspire/internal/domain/entity"
)

// SessionRepository defines session persistence operations
type SessionRepository interface {
	// Create creates a new session in the repository
	Create(ctx context.Context, session *entity.Session) error

	// GetByID retrieves a session by its ID
	GetByID(ctx context.Context, id string) (*entity.Session, error)

	// GetAll retrieves all sessions
	GetAll(ctx context.Context) ([]*entity.Session, error)

	// Update updates an existing session
	Update(ctx context.Context, session *entity.Session) error

	// Delete removes a session by its ID
	Delete(ctx context.Context, id string) error

	// UpdateStatus updates only the status of a session
	UpdateStatus(ctx context.Context, id string, status entity.Status) error
}

// EventHandler is a function that handles WhatsApp events
type EventHandler func(event *entity.Event)

// QREvent represents a QR code event during authentication
type QREvent struct {
	Type    string // "qr", "authenticated", "error", "timeout"
	Data    string // base64 QR image or JID or error message
	Message string // optional message
}

// NewQRCodeEvent creates a QR code event
func NewQRCodeEvent(base64Image string) QREvent {
	return QREvent{
		Type: "qr",
		Data: base64Image,
	}
}

// NewAuthenticatedEvent creates an authenticated event
func NewAuthenticatedEvent(jid string) QREvent {
	return QREvent{
		Type: "authenticated",
		Data: jid,
	}
}

// NewQRErrorEvent creates an error event
func NewQRErrorEvent(message string) QREvent {
	return QREvent{
		Type:    "error",
		Message: message,
	}
}

// NewQRTimeoutEvent creates a timeout event
func NewQRTimeoutEvent() QREvent {
	return QREvent{
		Type:    "timeout",
		Message: "QR authentication timed out",
	}
}

// WhatsAppClient defines WhatsApp operations
type WhatsAppClient interface {
	// Connect establishes a connection for the given session
	Connect(ctx context.Context, sessionID string) error

	// Disconnect closes the connection for the given session
	Disconnect(ctx context.Context, sessionID string) error

	// SendMessage sends a message through WhatsApp
	SendMessage(ctx context.Context, msg *entity.Message) error

	// SendReaction sends a reaction to a message
	SendReaction(ctx context.Context, sessionID, chatJID, messageID, emoji string) error

	// SendReadReceipt sends read receipts for multiple messages atomically
	SendReadReceipt(ctx context.Context, sessionID, chatJID string, messageIDs []string) error

	// SendPresence sends a presence update (typing, paused, online, offline)
	SendPresence(ctx context.Context, sessionID, chatJID, state string) error

	// GetQRChannel returns a channel that receives QR code events for authentication
	GetQRChannel(ctx context.Context, sessionID string) (<-chan QREvent, error)

	// RegisterEventHandler registers a handler for WhatsApp events
	RegisterEventHandler(handler EventHandler)

	// IsConnected checks if a session is currently connected
	IsConnected(sessionID string) bool

	// GetSessionJID returns the JID for a connected session
	GetSessionJID(sessionID string) (string, error)

	// SetSessionJIDMapping sets the JID mapping for a session (used for reconnection after restart)
	SetSessionJIDMapping(sessionID, jid string)

	// SetHistorySyncConfig sets the history sync configuration for a session
	SetHistorySyncConfig(sessionID string, enabled, fullSync bool, since string)

	// GetHistorySyncConfig gets the history sync configuration for a session
	GetHistorySyncConfig(sessionID string) (enabled, fullSync bool, since string)

	// CheckPhoneNumber checks if a phone number is registered on WhatsApp
	CheckPhoneNumber(ctx context.Context, sessionID, phone string) (*entity.Contact, error)

	// GetUserProfile retrieves the profile information for a user
	GetUserProfile(ctx context.Context, sessionID, jid string) (*entity.Contact, error)

	// ListContacts retrieves all contacts for a session
	ListContacts(ctx context.Context, sessionID string) ([]*entity.Contact, error)

	// ListChats retrieves all chats for a session
	ListChats(ctx context.Context, sessionID string) ([]*entity.Chat, error)
}

// EventPublisher defines event propagation operations to the API server
type EventPublisher interface {
	// Publish sends an event to the API server
	Publish(ctx context.Context, event *entity.Event) error

	// Connect establishes the WebSocket connection to the API server
	Connect(ctx context.Context) error

	// Disconnect closes the WebSocket connection
	Disconnect(ctx context.Context) error

	// IsConnected checks if the publisher is connected to the API server
	IsConnected() bool

	// QueueSize returns the number of events waiting to be sent
	QueueSize() int
}

// MessageQueue defines message queuing operations for rate limiting
type MessageQueue interface {
	// Enqueue adds a message to the queue
	Enqueue(ctx context.Context, msg *entity.Message) error

	// Dequeue retrieves and removes the next message from the queue
	Dequeue(ctx context.Context) (*entity.Message, error)

	// Size returns the number of messages in the queue
	Size() int

	// Clear removes all messages from the queue
	Clear()
}

// GroupFetcher defines operations for fetching groups from WhatsApp
type GroupFetcher interface {
	// GetJoinedGroups fetches all groups the session is a member of from WhatsApp
	GetJoinedGroups(ctx context.Context, sessionID string) ([]*entity.Group, error)
}

// EventRepository defines event persistence operations for debugging and audit
type EventRepository interface {
	// Create stores a new event in the repository
	Create(ctx context.Context, event *entity.Event) error

	// GetByID retrieves an event by its ID
	GetByID(ctx context.Context, id string) (*entity.Event, error)

	// List retrieves events with optional filters
	List(ctx context.Context, filter EventFilter) ([]*entity.Event, error)

	// DeleteOlderThan removes events older than the specified timestamp
	DeleteOlderThan(ctx context.Context, timestamp string) (int64, error)

	// Count returns the total number of events matching the filter
	Count(ctx context.Context, filter EventFilter) (int64, error)
}

// EventFilter defines filtering options for event queries
type EventFilter struct {
	SessionID *string           // Filter by session ID
	EventType *entity.EventType // Filter by event type
	Since     *string           // Filter events after this timestamp (RFC3339)
	Until     *string           // Filter events before this timestamp (RFC3339)
	Limit     int               // Maximum number of results (0 = no limit)
	Offset    int               // Number of results to skip
}

// WebhookConfigRepository defines webhook configuration persistence operations
type WebhookConfigRepository interface {
	// Create creates a new webhook configuration
	Create(ctx context.Context, config *entity.WebhookConfig) error

	// GetBySessionID retrieves webhook configuration for a session
	GetBySessionID(ctx context.Context, sessionID string) (*entity.WebhookConfig, error)

	// Update updates an existing webhook configuration
	Update(ctx context.Context, config *entity.WebhookConfig) error

	// Delete removes a webhook configuration by session ID
	Delete(ctx context.Context, sessionID string) error

	// Exists checks if a webhook configuration exists for a session
	Exists(ctx context.Context, sessionID string) (bool, error)
}
