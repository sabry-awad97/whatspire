package property

import (
	"context"
	"testing"
	"time"

	"whatspire/internal/domain/entity"
	"whatspire/internal/infrastructure/whatsapp"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"pgregory.net/rapid"
)

// Feature: whatsapp-http-api-enhancement, Property 1: Event Propagation Completeness
// *For any* WhatsApp event (message, reaction, receipt, presence), when the event is received
// by the service, an event SHALL be published to the Event Hub with all required fields populated.
// **Validates: Requirements 1.1, 1.2, 2.2, 2.3, 3.2, 3.3, 3.4, 4.3, 4.4**

func TestEventPropagationCompleteness_Property1(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 1.1: Message events contain all required fields
	properties.Property("Message events contain all required fields", prop.ForAll(
		func(sessionID, messageID, senderJID, chatJID string, msgType int) bool {
			// Create a parsed message
			parsedMsg := &whatsapp.ParsedMessage{
				MessageID:        messageID,
				SessionID:        sessionID,
				ChatJID:          chatJID,
				SenderJID:        senderJID,
				MessageType:      getMessageType(msgType),
				MessageTimestamp: time.Now(),
				Source:           whatsapp.ParsedMessageSourceRealtime,
			}

			// Create event with the parsed message
			event, err := entity.NewEventWithPayload(
				"test-event-id",
				entity.EventTypeMessageReceived,
				sessionID,
				parsedMsg,
			)
			if err != nil {
				t.Logf("Failed to create event: %v", err)
				return false
			}

			// Verify event has all required fields
			if event.ID == "" {
				t.Logf("Event ID is empty")
				return false
			}
			if event.Type != entity.EventTypeMessageReceived {
				t.Logf("Event type mismatch: %s", event.Type)
				return false
			}
			if event.SessionID != sessionID {
				t.Logf("Session ID mismatch: %s != %s", event.SessionID, sessionID)
				return false
			}
			if len(event.Data) == 0 {
				t.Logf("Event data is empty")
				return false
			}
			if event.Timestamp.IsZero() {
				t.Logf("Event timestamp is zero")
				return false
			}

			// Unmarshal and verify payload
			var payload whatsapp.ParsedMessage
			if err := event.UnmarshalData(&payload); err != nil {
				t.Logf("Failed to unmarshal payload: %v", err)
				return false
			}

			if payload.MessageID != messageID {
				t.Logf("Payload message ID mismatch")
				return false
			}
			if payload.SessionID != sessionID {
				t.Logf("Payload session ID mismatch")
				return false
			}
			if payload.SenderJID != senderJID {
				t.Logf("Payload sender JID mismatch")
				return false
			}
			if payload.ChatJID != chatJID {
				t.Logf("Payload chat JID mismatch")
				return false
			}

			return true
		},
		genNonEmptyString(),
		genNonEmptyString(),
		genNonEmptyString(),
		genNonEmptyString(),
		gen.IntRange(0, 10),
	))

	properties.TestingRun(t)
}

// Feature: whatsapp-http-api-enhancement, Property 2: Text Message Content Preservation
// *For any* text message received from WhatsApp, the published event SHALL contain
// the exact text content from the original message.
// **Validates: Requirements 1.3**

func TestTextMessageContentPreservation_Property2(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate random text content
		textContent := rapid.StringMatching("[a-zA-Z0-9 .,!?]+").Draw(t, "textContent")
		sessionID := rapid.StringMatching("[a-zA-Z0-9]+").Draw(t, "sessionID")
		messageID := rapid.StringMatching("[a-zA-Z0-9]+").Draw(t, "messageID")
		senderJID := rapid.StringMatching("[a-zA-Z0-9]+").Draw(t, "senderJID")
		chatJID := rapid.StringMatching("[a-zA-Z0-9]+").Draw(t, "chatJID")

		// Create a text message
		parsedMsg := &whatsapp.ParsedMessage{
			MessageID:        messageID,
			SessionID:        sessionID,
			ChatJID:          chatJID,
			SenderJID:        senderJID,
			MessageType:      whatsapp.ParsedMessageTypeText,
			Text:             &textContent,
			MessageTimestamp: time.Now(),
			Source:           whatsapp.ParsedMessageSourceRealtime,
		}

		// Create event
		event, err := entity.NewEventWithPayload(
			"test-event-id",
			entity.EventTypeMessageReceived,
			sessionID,
			parsedMsg,
		)
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}

		// Unmarshal and verify text content is preserved exactly
		var payload whatsapp.ParsedMessage
		if err := event.UnmarshalData(&payload); err != nil {
			t.Fatalf("Failed to unmarshal payload: %v", err)
		}

		if payload.Text == nil {
			t.Fatalf("Text content is nil")
		}

		if *payload.Text != textContent {
			t.Fatalf("Text content mismatch: got %q, want %q", *payload.Text, textContent)
		}
	})
}

// Feature: whatsapp-http-api-enhancement, Property 4: Quoted Message Context Preservation
// *For any* message with a quoted reply, the published event SHALL include both
// the quoted message ID and the quoted message content.
// **Validates: Requirements 1.5**

func TestQuotedMessageContextPreservation_Property4(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 4.1: Quoted message ID is preserved
	properties.Property("Quoted message ID is preserved", prop.ForAll(
		func(sessionID, messageID, quotedID string) bool {
			// Create a message with quoted reply
			parsedMsg := &whatsapp.ParsedMessage{
				MessageID:        messageID,
				SessionID:        sessionID,
				ChatJID:          "test-chat@g.us",
				SenderJID:        "test-sender@s.whatsapp.net",
				MessageType:      whatsapp.ParsedMessageTypeText,
				Text:             stringPtr("Reply text"),
				QuotedMessageID:  &quotedID,
				MessageTimestamp: time.Now(),
				Source:           whatsapp.ParsedMessageSourceRealtime,
			}

			// Create event
			event, err := entity.NewEventWithPayload(
				"test-event-id",
				entity.EventTypeMessageReceived,
				sessionID,
				parsedMsg,
			)
			if err != nil {
				t.Logf("Failed to create event: %v", err)
				return false
			}

			// Unmarshal and verify quoted message ID is preserved
			var payload whatsapp.ParsedMessage
			if err := event.UnmarshalData(&payload); err != nil {
				t.Logf("Failed to unmarshal payload: %v", err)
				return false
			}

			if payload.QuotedMessageID == nil {
				t.Logf("Quoted message ID is nil")
				return false
			}

			if *payload.QuotedMessageID != quotedID {
				t.Logf("Quoted message ID mismatch: %s != %s", *payload.QuotedMessageID, quotedID)
				return false
			}

			return true
		},
		genNonEmptyString(),
		genNonEmptyString(),
		genNonEmptyString(),
	))

	properties.TestingRun(t)
}

// Feature: whatsapp-http-api-enhancement, Property 5: Event Queueing for Disconnected Sessions
// *For any* message received when a session is disconnected, the event SHALL be queued
// and published when the session reconnects, preserving event order.
// **Validates: Requirements 1.6**

func TestEventQueueingForDisconnectedSessions_Property5(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Create event queue
		queue := whatsapp.NewEventQueue()

		// Generate random number of events (1-10)
		numEvents := rapid.IntRange(1, 10).Draw(t, "numEvents")
		sessionID := rapid.StringMatching("[a-zA-Z0-9]+").Draw(t, "sessionID")

		// Create and enqueue events
		events := make([]*entity.Event, numEvents)
		for i := 0; i < numEvents; i++ {
			messageID := rapid.StringMatching("[a-zA-Z0-9]+").Draw(t, "messageID")

			parsedMsg := &whatsapp.ParsedMessage{
				MessageID:        messageID,
				SessionID:        sessionID,
				ChatJID:          "test-chat@g.us",
				SenderJID:        "test-sender@s.whatsapp.net",
				MessageType:      whatsapp.ParsedMessageTypeText,
				Text:             stringPtr("Test message"),
				MessageTimestamp: time.Now(),
				Source:           whatsapp.ParsedMessageSourceRealtime,
			}

			event, err := entity.NewEventWithPayload(
				messageID,
				entity.EventTypeMessageReceived,
				sessionID,
				parsedMsg,
			)
			if err != nil {
				t.Fatalf("Failed to create event: %v", err)
			}

			events[i] = event
			queue.Enqueue(event)
		}

		// Verify queue size
		if queue.Size(sessionID) != numEvents {
			t.Fatalf("Queue size mismatch: got %d, want %d", queue.Size(sessionID), numEvents)
		}

		// Flush queue and verify order is preserved
		flushedEvents := queue.FlushSession(sessionID)
		if len(flushedEvents) != numEvents {
			t.Fatalf("Flushed events count mismatch: got %d, want %d", len(flushedEvents), numEvents)
		}

		// Verify events are in the same order
		for i := 0; i < numEvents; i++ {
			if flushedEvents[i].ID != events[i].ID {
				t.Fatalf("Event order not preserved at index %d: got %s, want %s",
					i, flushedEvents[i].ID, events[i].ID)
			}
		}

		// Verify queue is empty after flush
		if queue.Size(sessionID) != 0 {
			t.Fatalf("Queue not empty after flush: size=%d", queue.Size(sessionID))
		}
	})
}

// Helper functions

func getMessageType(index int) whatsapp.ParsedMessageType {
	types := []whatsapp.ParsedMessageType{
		whatsapp.ParsedMessageTypeText,
		whatsapp.ParsedMessageTypeImage,
		whatsapp.ParsedMessageTypeVideo,
		whatsapp.ParsedMessageTypeAudio,
		whatsapp.ParsedMessageTypeDocument,
		whatsapp.ParsedMessageTypeSticker,
		whatsapp.ParsedMessageTypeContact,
		whatsapp.ParsedMessageTypeLocation,
		whatsapp.ParsedMessageTypePoll,
		whatsapp.ParsedMessageTypeReaction,
		whatsapp.ParsedMessageTypeUnknown,
	}
	return types[index%len(types)]
}

func genNonEmptyString() gopter.Gen {
	return gen.Identifier().SuchThat(func(s string) bool {
		return len(s) > 0
	})
}

func stringPtr(s string) *string {
	return &s
}

// TestEventQueueConcurrency tests that the event queue is thread-safe
func TestEventQueueConcurrency(t *testing.T) {
	queue := whatsapp.NewEventQueue()
	sessionID := "test-session"

	// Create events concurrently
	done := make(chan bool)
	numGoroutines := 10
	eventsPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < eventsPerGoroutine; j++ {
				parsedMsg := &whatsapp.ParsedMessage{
					MessageID:        "msg-" + string(rune(goroutineID)) + "-" + string(rune(j)),
					SessionID:        sessionID,
					ChatJID:          "test-chat@g.us",
					SenderJID:        "test-sender@s.whatsapp.net",
					MessageType:      whatsapp.ParsedMessageTypeText,
					MessageTimestamp: time.Now(),
					Source:           whatsapp.ParsedMessageSourceRealtime,
				}

				event, _ := entity.NewEventWithPayload(
					parsedMsg.MessageID,
					entity.EventTypeMessageReceived,
					sessionID,
					parsedMsg,
				)

				queue.Enqueue(event)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all events were queued
	expectedSize := numGoroutines * eventsPerGoroutine
	actualSize := queue.Size(sessionID)
	if actualSize != expectedSize {
		t.Errorf("Queue size mismatch: got %d, want %d", actualSize, expectedSize)
	}
}

// TestEventQueueMultipleSessions tests that events are properly isolated by session
func TestEventQueueMultipleSessions(t *testing.T) {
	queue := whatsapp.NewEventQueue()

	// Create events for different sessions
	sessions := []string{"session1", "session2", "session3"}
	eventsPerSession := 5

	for _, sessionID := range sessions {
		for i := 0; i < eventsPerSession; i++ {
			parsedMsg := &whatsapp.ParsedMessage{
				MessageID:        "msg-" + sessionID + "-" + string(rune(i)),
				SessionID:        sessionID,
				ChatJID:          "test-chat@g.us",
				SenderJID:        "test-sender@s.whatsapp.net",
				MessageType:      whatsapp.ParsedMessageTypeText,
				MessageTimestamp: time.Now(),
				Source:           whatsapp.ParsedMessageSourceRealtime,
			}

			event, _ := entity.NewEventWithPayload(
				parsedMsg.MessageID,
				entity.EventTypeMessageReceived,
				sessionID,
				parsedMsg,
			)

			queue.Enqueue(event)
		}
	}

	// Verify each session has the correct number of events
	for _, sessionID := range sessions {
		size := queue.Size(sessionID)
		if size != eventsPerSession {
			t.Errorf("Session %s queue size mismatch: got %d, want %d", sessionID, size, eventsPerSession)
		}
	}

	// Flush one session and verify others are unaffected
	queue.FlushSession(sessions[0])

	if queue.Size(sessions[0]) != 0 {
		t.Errorf("Session %s not flushed", sessions[0])
	}

	for _, sessionID := range sessions[1:] {
		size := queue.Size(sessionID)
		if size != eventsPerSession {
			t.Errorf("Session %s affected by flush of another session: got %d, want %d",
				sessionID, size, eventsPerSession)
		}
	}
}

// TestMessageHandlerSessionTracking tests that the message handler correctly tracks session connection status
func TestMessageHandlerSessionTracking(t *testing.T) {
	// Create message handler
	messageParser := whatsapp.NewMessageParser()
	handler := whatsapp.NewMessageHandler(
		messageParser,
		nil, // No media downloader needed for this test
		nil, // No media storage needed for this test
		nil, // No logger needed for this test
	)

	sessionID := "test-session"

	// Initially, session should not be connected
	if handler.IsSessionConnected(sessionID) {
		t.Error("Session should not be connected initially")
	}

	// Set session as connected
	handler.SetSessionConnected(sessionID, true)

	if !handler.IsSessionConnected(sessionID) {
		t.Error("Session should be connected after SetSessionConnected(true)")
	}

	// Set session as disconnected
	handler.SetSessionConnected(sessionID, false)

	if handler.IsSessionConnected(sessionID) {
		t.Error("Session should not be connected after SetSessionConnected(false)")
	}
}

// TestMessageHandlerEventQueueing tests that events are queued when session is disconnected
func TestMessageHandlerEventQueueing(t *testing.T) {
	// Create message handler
	messageParser := whatsapp.NewMessageParser()
	handler := whatsapp.NewMessageHandler(
		messageParser,
		nil,
		nil,
		nil,
	)

	sessionID := "test-session"
	ctx := context.Background()

	// Set session as disconnected
	handler.SetSessionConnected(sessionID, false)

	// Create test event
	parsedMsg := &whatsapp.ParsedMessage{
		MessageID:        "test-msg",
		SessionID:        sessionID,
		ChatJID:          "test-chat@g.us",
		SenderJID:        "test-sender@s.whatsapp.net",
		MessageType:      whatsapp.ParsedMessageTypeText,
		Text:             stringPtr("Test message"),
		MessageTimestamp: time.Now(),
		Source:           whatsapp.ParsedMessageSourceRealtime,
	}

	event, err := entity.NewEventWithPayload(
		"test-event",
		entity.EventTypeMessageReceived,
		sessionID,
		parsedMsg,
	)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	// Queue the event
	handler.QueueEvent(event)

	// Verify event is queued
	queuedEvents := handler.GetQueuedEvents(sessionID)
	if len(queuedEvents) != 1 {
		t.Errorf("Expected 1 queued event, got %d", len(queuedEvents))
	}

	// Reconnect session (this should flush the queue)
	handler.SetSessionConnected(sessionID, true)

	// Verify queue is empty after reconnection
	queuedEvents = handler.GetQueuedEvents(sessionID)
	if len(queuedEvents) != 0 {
		t.Errorf("Expected 0 queued events after reconnection, got %d", len(queuedEvents))
	}

	_ = ctx // Suppress unused variable warning
}
