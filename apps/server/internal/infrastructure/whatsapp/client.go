package whatsapp

import (
	"context"
	"strings"
	"sync"
	"time"

	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/errors"
	"whatspire/internal/domain/repository"

	"go.mau.fi/whatsmeow"
	waCompanionReg "go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	_ "modernc.org/sqlite" // SQLite driver for whatsmeow store
)

// ClientConfig holds configuration for the WhatsApp client
type ClientConfig struct {
	DBPath           string
	QRTimeout        time.Duration
	ReconnectDelay   time.Duration
	MaxReconnects    int
	MessageRateLimit int
	// Circuit breaker configuration
	CircuitBreakerEnabled bool
	CircuitBreakerConfig  CircuitBreakerConfig
}

// DefaultClientConfig returns default configuration
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		DBPath:                "/data/whatsapp.db",
		QRTimeout:             2 * time.Minute,
		ReconnectDelay:        5 * time.Second,
		MaxReconnects:         10,
		MessageRateLimit:      30,
		CircuitBreakerEnabled: true,
		CircuitBreakerConfig:  DefaultCircuitBreakerConfig(),
	}
}

// WhatsmeowClient implements WhatsAppClient using whatsmeow
type WhatsmeowClient struct {
	config          ClientConfig
	container       *sqlstore.Container
	clients         map[string]*whatsmeow.Client
	sessionToJID    map[string]string // Maps session UUID to WhatsApp JID user part
	mu              sync.RWMutex
	handlers        []repository.EventHandler
	logger          waLog.Logger
	circuitBreaker  *CircuitBreaker
	mediaUploader   *WhatsmeowMediaUploader
	messageParser   *MessageParser
	messageHandler  *MessageHandler
	reactionHandler *ReactionHandler
	presenceRepo    repository.PresenceRepository

	// History sync configuration per session
	historySyncConfig map[string]HistorySyncConfig
	historySyncMu     sync.RWMutex
}

// HistorySyncConfig holds history sync configuration for a session
type HistorySyncConfig struct {
	Enabled  bool
	FullSync bool
	Since    string // ISO 8601 timestamp for incremental sync
}

// NewWhatsmeowClient creates a new WhatsApp client
func NewWhatsmeowClient(ctx context.Context, config ClientConfig) (*WhatsmeowClient, error) {
	// Create a logger
	logger := waLog.Stdout("WhatsApp", "INFO", true)

	// Set device properties for the linked device name shown in WhatsApp
	store.DeviceProps.Os = proto.String("PharmaBroker")
	store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_DESKTOP.Enum()
	store.DeviceProps.RequireFullSync = proto.Bool(true) // Sync group chats and history

	// Create the SQL store container with foreign keys enabled (required by whatsmeow)
	// Using modernc.org/sqlite pragma syntax: _pragma=foreign_keys(1)
	dsn := config.DBPath + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	container, err := sqlstore.New(ctx, "sqlite", dsn, logger)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithCause(err).WithMessage("failed to create whatsmeow store")
	}

	client := &WhatsmeowClient{
		config:            config,
		container:         container,
		clients:           make(map[string]*whatsmeow.Client),
		sessionToJID:      make(map[string]string),
		handlers:          make([]repository.EventHandler, 0),
		logger:            logger,
		messageParser:     NewMessageParser(),
		historySyncConfig: make(map[string]HistorySyncConfig),
	}

	// Initialize circuit breaker if enabled
	if config.CircuitBreakerEnabled {
		client.circuitBreaker = NewCircuitBreaker(config.CircuitBreakerConfig)
	}

	return client, nil
}

// RegisterEventHandler registers a handler for WhatsApp events
func (c *WhatsmeowClient) RegisterEventHandler(handler repository.EventHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers = append(c.handlers, handler)
}

// IsConnected checks if a session is currently connected
func (c *WhatsmeowClient) IsConnected(sessionID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	client, exists := c.clients[sessionID]
	if !exists {
		return false
	}
	return client.IsConnected()
}

// SetSessionJIDMapping sets the JID mapping for a session (used for reconnection)
func (c *WhatsmeowClient) SetSessionJIDMapping(sessionID, jid string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Extract user part from JID (e.g., "201021347532" from "201021347532:123@s.whatsapp.net")
	jidUser := jid
	if idx := strings.Index(jid, ":"); idx > 0 {
		jidUser = jid[:idx]
	} else if idx := strings.Index(jid, "@"); idx > 0 {
		jidUser = jid[:idx]
	}

	c.sessionToJID[sessionID] = jidUser
	c.logger.Infof("SetSessionJIDMapping: sessionID=%s, jidUser=%s", sessionID, jidUser)
}

// GetSessionJID returns the JID for a connected session
func (c *WhatsmeowClient) GetSessionJID(sessionID string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	client, exists := c.clients[sessionID]
	if !exists {
		return "", errors.ErrSessionNotFound
	}

	if client.Store.ID == nil {
		return "", errors.ErrSessionInvalid
	}

	return client.Store.ID.String(), nil
}

// Close closes all connections and the container
func (c *WhatsmeowClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, client := range c.clients {
		client.Disconnect()
	}
	c.clients = make(map[string]*whatsmeow.Client)

	return nil
}

// GetStoredSessions returns all session IDs that have stored credentials in whatsmeow's database
// These sessions can be auto-reconnected without QR scan
func (c *WhatsmeowClient) GetStoredSessions(ctx context.Context) ([]string, error) {
	devices, err := c.container.GetAllDevices(ctx)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithCause(err)
	}

	sessionIDs := make([]string, 0, len(devices))
	for _, device := range devices {
		if device.ID != nil {
			sessionIDs = append(sessionIDs, device.ID.User)
		}
	}
	return sessionIDs, nil
}

// AutoReconnect attempts to reconnect all sessions that have stored credentials
// GetCircuitBreakerState returns the current state of the circuit breaker
// Returns empty string if circuit breaker is not enabled
func (c *WhatsmeowClient) GetCircuitBreakerState() string {
	if c.circuitBreaker == nil {
		return ""
	}
	return c.circuitBreaker.State().String()
}

// IsCircuitBreakerOpen returns true if the circuit breaker is open
func (c *WhatsmeowClient) IsCircuitBreakerOpen() bool {
	if c.circuitBreaker == nil {
		return false
	}
	return c.circuitBreaker.IsOpen()
}

// SetMediaUploader sets the media uploader for handling media messages
func (c *WhatsmeowClient) SetMediaUploader(uploader *WhatsmeowMediaUploader) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mediaUploader = uploader
}

// SetMessageHandler sets the message handler for processing incoming messages
func (c *WhatsmeowClient) SetMessageHandler(handler *MessageHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messageHandler = handler
}

// SetReactionHandler sets the reaction handler for processing incoming reactions
func (c *WhatsmeowClient) SetReactionHandler(handler *ReactionHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reactionHandler = handler
}

// SetPresenceRepository sets the presence repository for storing presence updates
func (c *WhatsmeowClient) SetPresenceRepository(repo repository.PresenceRepository) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.presenceRepo = repo
}

// getOrCreateDevice gets or creates a device store for the session
func (c *WhatsmeowClient) getOrCreateDevice(ctx context.Context, sessionID string) (*store.Device, error) {
	// Try to get existing device
	devices, err := c.container.GetAllDevices(ctx)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithCause(err)
	}

	// First, check if we have a JID mapping for this session
	if jidUser, ok := c.sessionToJID[sessionID]; ok {
		for _, device := range devices {
			if device.ID != nil && device.ID.User == jidUser {
				c.logger.Infof("getOrCreateDevice: found device by JID mapping: sessionID=%s, jidUser=%s", sessionID, jidUser)
				return device, nil
			}
		}
	}

	// Look for device with matching session ID in JID (legacy check)
	for _, device := range devices {
		if device.ID != nil && device.ID.User == sessionID {
			return device, nil
		}
	}

	// Create new device
	c.logger.Infof("getOrCreateDevice: creating new device for sessionID=%s", sessionID)
	device := c.container.NewDevice()
	return device, nil
}

// handleEvent processes WhatsApp events and dispatches to handlers
func (c *WhatsmeowClient) handleEvent(sessionID string, client *whatsmeow.Client, evt interface{}) {
	var event *entity.Event
	var err error

	switch v := evt.(type) {
	case *events.Message:
		event, err = c.handleMessageEvent(sessionID, client, v)
	case *events.Connected:
		// Notify message handler of connection
		if c.messageHandler != nil {
			c.messageHandler.SetSessionConnected(sessionID, true)
		}
		event, err = entity.NewEventWithPayload(
			generateEventID(),
			entity.EventTypeConnected,
			sessionID,
			map[string]string{"status": "connected"},
		)
	case *events.Disconnected:
		// Notify message handler of disconnection
		if c.messageHandler != nil {
			c.messageHandler.SetSessionConnected(sessionID, false)
		}
		event, err = entity.NewEventWithPayload(
			generateEventID(),
			entity.EventTypeDisconnected,
			sessionID,
			map[string]string{"status": "disconnected"},
		)
	case *events.LoggedOut:
		event, err = entity.NewEventWithPayload(
			generateEventID(),
			entity.EventTypeLoggedOut,
			sessionID,
			map[string]string{"reason": "logged_out"},
		)
		// Remove client from map
		c.mu.Lock()
		delete(c.clients, sessionID)
		c.mu.Unlock()
	case *events.Receipt:
		event, err = c.handleReceiptEvent(sessionID, v)
	case *events.Presence:
		event, err = c.handlePresenceEvent(sessionID, v)
	case *events.HistorySync:
		// Handle history sync to extract pushnames and emit message events
		c.handleHistorySyncEvent(sessionID, client, v)
		return // Don't emit domain event for history sync itself
	default:
		// Ignore other events
		return
	}

	if err != nil || event == nil {
		return
	}

	// Dispatch to all handlers
	c.mu.RLock()
	handlers := make([]repository.EventHandler, len(c.handlers))
	copy(handlers, c.handlers)
	c.mu.RUnlock()

	for _, handler := range handlers {
		handler(event)
	}
}

// handleMessageEvent converts a WhatsApp message event to a domain event
// Also saves the sender's pushname to the contact store for future lookups
func (c *WhatsmeowClient) handleMessageEvent(sessionID string, client *whatsmeow.Client, msg *events.Message) (*entity.Event, error) {
	// Use the message handler if available (supports media download and queueing)
	if c.messageHandler != nil {
		ctx := context.Background()
		event, err := c.messageHandler.HandleIncomingMessage(ctx, sessionID, client, msg)
		if err != nil {
			c.logger.Warnf("Message handler failed: %v", err)
			return nil, err
		}

		// Check if session is connected, queue event if not
		if !c.messageHandler.IsSessionConnected(sessionID) {
			c.logger.Infof("Session %s disconnected, queueing event", sessionID)
			c.messageHandler.QueueEvent(event)
			return nil, nil // Don't emit event yet
		}

		return event, nil
	}

	// Fallback to legacy handling if message handler not set
	// Save pushname to contact store if available
	if msg.Info.PushName != "" && client != nil && client.Store != nil && client.Store.Contacts != nil {
		// Update contact with pushname (use background context since this is async)
		ctx := context.Background()
		_, _, err := client.Store.Contacts.PutPushName(ctx, msg.Info.Sender, msg.Info.PushName)
		if err != nil {
			c.logger.Warnf("Failed to save pushname for %s: %v", msg.Info.Sender.String(), err)
		}
	}

	// Use the message parser to create a full ParsedMessage
	parsedMsg, err := c.messageParser.ParseRealtimeMessage(sessionID, msg)
	if err != nil {
		c.logger.Warnf("Failed to parse message: %v", err)
		return nil, err
	}

	// Resolve LID to phone number JID if needed
	if msg.Info.Sender.Server == "lid" && client != nil && client.Store != nil && client.Store.LIDs != nil {
		ctx := context.Background()
		pnJID, err := client.Store.LIDs.GetPNForLID(ctx, msg.Info.Sender)
		if err == nil && !pnJID.IsEmpty() {
			parsedMsg.SenderJID = pnJID.String()
			c.logger.Debugf("Resolved LID %s to PN %s", msg.Info.Sender.String(), pnJID.String())
		}
	}

	// Emit the full parsed message as the event payload
	return entity.NewEventWithPayload(
		generateEventID(),
		entity.EventTypeMessageReceived,
		sessionID,
		parsedMsg,
	)
}

// handleReceiptEvent converts a WhatsApp receipt event to a domain event
func (c *WhatsmeowClient) handleReceiptEvent(sessionID string, receipt *events.Receipt) (*entity.Event, error) {
	var eventType entity.EventType

	switch receipt.Type {
	case types.ReceiptTypeDelivered:
		eventType = entity.EventTypeMessageDelivered
	case types.ReceiptTypeRead:
		eventType = entity.EventTypeMessageRead
	default:
		return nil, nil // Ignore other receipt types
	}

	payload := map[string]interface{}{
		"message_ids": receipt.MessageIDs,
		"from":        receipt.MessageSource.Sender.String(),
		"timestamp":   receipt.Timestamp,
	}

	return entity.NewEventWithPayload(
		generateEventID(),
		eventType,
		sessionID,
		payload,
	)
}

// handlePresenceEvent converts a WhatsApp presence event to a domain event
func (c *WhatsmeowClient) handlePresenceEvent(sessionID string, presence *events.Presence) (*entity.Event, error) {
	// Map whatsmeow presence to our domain presence state
	var state entity.PresenceState

	// Check if it's unavailable (offline)
	if presence.Unavailable {
		state = entity.PresenceStateOffline
	} else {
		// Available means online
		state = entity.PresenceStateOnline
	}

	// Create presence entity
	presenceEntity := entity.NewPresence(
		generateEventID(),
		sessionID,
		presence.From.String(),
		"", // Chat JID is empty for general presence
		state,
	)

	// Save to repository if available
	if c.presenceRepo != nil {
		ctx := context.Background()
		if err := c.presenceRepo.Save(ctx, presenceEntity); err != nil {
			c.logger.Warnf("Failed to save presence: %v", err)
		}
	}

	return entity.NewEventWithPayload(
		generateEventID(),
		entity.EventTypePresenceUpdate,
		sessionID,
		presenceEntity,
	)
}

// handleHistorySyncEvent processes history sync events to extract and save pushnames
// This is called during initial connection when WhatsApp syncs message history
// It also emits message events for each message in the history
func (c *WhatsmeowClient) handleHistorySyncEvent(sessionID string, client *whatsmeow.Client, evt *events.HistorySync) {
	if client == nil || client.Store == nil {
		return
	}

	// Check if history sync is enabled for this session
	enabled, fullSync, sinceStr := c.GetHistorySyncConfig(sessionID)
	if !enabled {
		c.logger.Infof("History sync disabled for session %s, skipping", sessionID)
		return
	}

	ctx := context.Background()
	savedCount := 0
	messageCount := 0
	droppedCount := 0

	// Parse the "since" timestamp for incremental sync
	var sinceTime time.Time
	if !fullSync && sinceStr != "" {
		var err error
		sinceTime, err = time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			c.logger.Warnf("Failed to parse since timestamp '%s': %v, falling back to full sync", sinceStr, err)
			fullSync = true
		} else {
			c.logger.Infof("Incremental sync: filtering messages since %s", sinceTime.Format(time.RFC3339))
		}
	}

	// Get handlers for emitting events
	c.mu.RLock()
	handlers := make([]repository.EventHandler, len(c.handlers))
	copy(handlers, c.handlers)
	c.mu.RUnlock()

	// Process conversations from history sync
	if evt.Data != nil && evt.Data.Conversations != nil {
		for _, conv := range evt.Data.Conversations {
			// Extract pushname from conversation metadata
			if conv.Name != nil && *conv.Name != "" && conv.ID != nil {
				// Parse JID from conversation ID
				jid, err := types.ParseJID(*conv.ID)
				if err == nil && !jid.IsEmpty() && client.Store.Contacts != nil {
					_, _, err := client.Store.Contacts.PutPushName(ctx, jid, *conv.Name)
					if err == nil {
						savedCount++
					}
				}
			}

			// Get chat JID for this conversation
			chatJID := ""
			if conv.ID != nil {
				chatJID = *conv.ID
			}

			// Process messages in the conversation
			if conv.Messages != nil {
				for _, histMsg := range conv.Messages {
					if histMsg.Message == nil || histMsg.Message.Message == nil {
						continue
					}

					// Get message timestamp
					msgTimestamp := time.Unix(int64(histMsg.Message.GetMessageTimestamp()), 0)

					// Filter by timestamp for incremental sync
					if !fullSync && !sinceTime.IsZero() && msgTimestamp.Before(sinceTime) {
						droppedCount++
						continue
					}

					// Get sender JID and pushname from message
					var senderJID types.JID
					var pushName string

					if histMsg.Message.Key != nil && histMsg.Message.Key.RemoteJID != nil {
						jid, err := types.ParseJID(*histMsg.Message.Key.RemoteJID)
						if err == nil {
							senderJID = jid
						}
					}

					if histMsg.Message.PushName != nil && *histMsg.Message.PushName != "" {
						pushName = *histMsg.Message.PushName
					}

					// Save pushname if we have both JID and pushname
					if !senderJID.IsEmpty() && pushName != "" && senderJID.Server == types.DefaultUserServer && client.Store.Contacts != nil {
						_, _, err := client.Store.Contacts.PutPushName(ctx, senderJID, pushName)
						if err == nil {
							savedCount++
						}
					}

					// Build MessageInfo from history message
					var msgInfo *types.MessageInfo
					if histMsg.Message.Key != nil {
						msgInfo = &types.MessageInfo{
							ID:        types.MessageID(histMsg.Message.Key.GetID()),
							Timestamp: msgTimestamp,
							MessageSource: types.MessageSource{
								Sender:   senderJID,
								IsFromMe: histMsg.Message.Key.GetFromMe(),
							},
							PushName: pushName,
						}
					}

					// Parse the history message
					parsedMsg, err := c.messageParser.ParseHistoryMessage(sessionID, chatJID, histMsg.Message.Message, msgInfo)
					if err != nil {
						c.logger.Warnf("Failed to parse history message: %v", err)
						continue
					}

					// Resolve LID to phone number JID if needed
					if !senderJID.IsEmpty() && senderJID.Server == "lid" && client.Store != nil && client.Store.LIDs != nil {
						pnJID, err := client.Store.LIDs.GetPNForLID(ctx, senderJID)
						if err == nil && !pnJID.IsEmpty() {
							parsedMsg.SenderJID = pnJID.String()
						}
					}

					// Only emit events for group messages (as per requirements)
					if !parsedMsg.IsGroupMessage() {
						continue
					}

					// Create and emit the event
					event, err := entity.NewEventWithPayload(
						generateEventID(),
						entity.EventTypeMessageReceived,
						sessionID,
						parsedMsg,
					)
					if err != nil {
						c.logger.Warnf("Failed to create history message event: %v", err)
						continue
					}

					// Dispatch to all handlers
					for _, handler := range handlers {
						handler(event)
					}
					messageCount++
				}
			}
		}
	}

	// Log sync statistics
	syncType := "full"
	if !fullSync {
		syncType = "incremental"
	}

	if savedCount > 0 {
		c.logger.Infof("History sync (%s): saved %d pushnames to contact store", syncType, savedCount)
	}
	if messageCount > 0 {
		c.logger.Infof("History sync (%s): emitted %d message events", syncType, messageCount)
	}
	if droppedCount > 0 {
		c.logger.Infof("History sync (%s): dropped %d messages (before %s)", syncType, droppedCount, sinceTime.Format(time.RFC3339))
	}

	// Emit sync progress event
	c.emitSyncProgressEvent(sessionID, messageCount, messageCount+droppedCount)
}

// emitSyncProgressEvent emits a sync progress event to track history sync progress
func (c *WhatsmeowClient) emitSyncProgressEvent(sessionID string, stored, total int) {
	c.mu.RLock()
	handlers := make([]repository.EventHandler, len(c.handlers))
	copy(handlers, c.handlers)
	c.mu.RUnlock()

	event, err := entity.NewEventWithPayload(
		generateEventID(),
		entity.EventTypeSyncProgress,
		sessionID,
		map[string]interface{}{
			"stored":  stored,
			"dropped": total - stored,
			"total":   total,
		},
	)
	if err != nil {
		c.logger.Warnf("Failed to create sync progress event: %v", err)
		return
	}

	for _, handler := range handlers {
		handler(event)
	}
}

// SetHistorySyncConfig sets the history sync configuration for a session
func (c *WhatsmeowClient) SetHistorySyncConfig(sessionID string, enabled, fullSync bool, since string) {
	c.historySyncMu.Lock()
	defer c.historySyncMu.Unlock()

	c.historySyncConfig[sessionID] = HistorySyncConfig{
		Enabled:  enabled,
		FullSync: fullSync,
		Since:    since,
	}

	c.logger.Infof("SetHistorySyncConfig: sessionID=%s, enabled=%v, fullSync=%v, since=%s",
		sessionID, enabled, fullSync, since)
}

// GetHistorySyncConfig gets the history sync configuration for a session
func (c *WhatsmeowClient) GetHistorySyncConfig(sessionID string) (enabled, fullSync bool, since string) {
	c.historySyncMu.RLock()
	defer c.historySyncMu.RUnlock()

	config, exists := c.historySyncConfig[sessionID]
	if !exists {
		// Default: history sync disabled
		return false, false, ""
	}

	return config.Enabled, config.FullSync, config.Since
}
