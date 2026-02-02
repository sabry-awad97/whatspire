package property

import (
	"context"
	"testing"

	"whatspire/internal/application/usecase"
	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/errors"
	"whatspire/internal/infrastructure/persistence"
	"whatspire/test/mocks"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// ==================== Use Shared Mocks ====================

type WhatsAppClientMock = mocks.WhatsAppClientMock
type EventPublisherMock = mocks.EventPublisherMock

var (
	NewWhatsAppClientMock = mocks.NewWhatsAppClientMock
	NewEventPublisherMock = mocks.NewEventPublisherMock
)

// Feature: session-connection-flow, Property 3: Go Service Publishes Connecting Event on Connection Start
// *For any* session reconnection attempt in the Go service, a 'connection.connecting' event
// SHALL be published via WebSocket before the WhatsApp client connection is attempted.
// **Validates: Requirements 1.3, 2.2**

func TestConnectingEventPublication_Property3(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 3.1: ReconnectSessionWithJID publishes connection.connecting event
	properties.Property("reconnect publishes connecting event before connection attempt", prop.ForAll(
		func(sessionID string) bool {
			if sessionID == "" {
				return true // skip empty inputs
			}

			ctx := context.Background()
			db := setupTestDB(t)
			repo := persistence.NewSessionRepository(db)
			waClient := NewWhatsAppClientMock()
			publisher := NewEventPublisherMock()

			// Create session
			session := entity.NewSession(sessionID, "Test Session")
			_ = repo.Create(ctx, session)

			uc := usecase.NewSessionUseCase(repo, waClient, publisher, nil)

			// Track connection order
			var connectingEventPublished bool
			var connectCalled bool
			var connectingBeforeConnect bool

			// Set up mock to track order
			waClient.ConnectFn = func(ctx context.Context, sid string) error {
				connectCalled = true
				// Check if connecting event was published before Connect was called
				for _, event := range publisher.GetEvents() {
					if event.Type == entity.EventTypeConnectionConnecting && event.SessionID == sessionID {
						connectingEventPublished = true
						connectingBeforeConnect = true
					}
				}
				return nil
			}

			// Call reconnect
			_ = uc.ReconnectSessionWithJID(ctx, sessionID, "")

			// Verify connecting event was published
			for _, event := range publisher.GetEvents() {
				if event.Type == entity.EventTypeConnectionConnecting && event.SessionID == sessionID {
					connectingEventPublished = true
				}
			}

			return connectingEventPublished && connectCalled && connectingBeforeConnect
		},
		gen.Identifier(),
	))

	// Property 3.2: Connecting event has correct session_id
	properties.Property("connecting event has correct session_id", prop.ForAll(
		func(sessionID string) bool {
			if sessionID == "" {
				return true // skip empty inputs
			}

			ctx := context.Background()
			db := setupTestDB(t)
			repo := persistence.NewSessionRepository(db)
			waClient := NewWhatsAppClientMock()
			publisher := NewEventPublisherMock()

			// Create session
			session := entity.NewSession(sessionID, "Test Session")
			_ = repo.Create(ctx, session)

			uc := usecase.NewSessionUseCase(repo, waClient, publisher, nil)

			// Call reconnect
			_ = uc.ReconnectSessionWithJID(ctx, sessionID, "")

			// Find connecting event
			for _, event := range publisher.GetEvents() {
				if event.Type == entity.EventTypeConnectionConnecting {
					return event.SessionID == sessionID
				}
			}

			return false // No connecting event found
		},
		gen.Identifier(),
	))

	// Property 3.3: Connecting event has valid timestamp
	properties.Property("connecting event has valid timestamp", prop.ForAll(
		func(sessionID string) bool {
			if sessionID == "" {
				return true // skip empty inputs
			}

			ctx := context.Background()
			db := setupTestDB(t)
			repo := persistence.NewSessionRepository(db)
			waClient := NewWhatsAppClientMock()
			publisher := NewEventPublisherMock()

			// Create session
			session := entity.NewSession(sessionID, "Test Session")
			_ = repo.Create(ctx, session)

			uc := usecase.NewSessionUseCase(repo, waClient, publisher, nil)

			// Call reconnect
			_ = uc.ReconnectSessionWithJID(ctx, sessionID, "")

			// Find connecting event
			for _, event := range publisher.GetEvents() {
				if event.Type == entity.EventTypeConnectionConnecting {
					return !event.Timestamp.IsZero()
				}
			}

			return false // No connecting event found
		},
		gen.Identifier(),
	))

	// Property 3.4: Connecting event is published even when connection fails
	properties.Property("connecting event published even on connection failure", prop.ForAll(
		func(sessionID string) bool {
			if sessionID == "" {
				return true // skip empty inputs
			}

			ctx := context.Background()
			db := setupTestDB(t)
			repo := persistence.NewSessionRepository(db)
			waClient := NewWhatsAppClientMock()
			publisher := NewEventPublisherMock()

			// Create session
			session := entity.NewSession(sessionID, "Test Session")
			_ = repo.Create(ctx, session)

			// Make connection fail
			waClient.ConnectFn = func(ctx context.Context, sid string) error {
				return errors.ErrConnectionFailed.WithMessage("simulated failure")
			}

			uc := usecase.NewSessionUseCase(repo, waClient, publisher, nil)

			// Call reconnect (will fail)
			_ = uc.ReconnectSessionWithJID(ctx, sessionID, "")

			// Verify connecting event was still published
			for _, event := range publisher.GetEvents() {
				if event.Type == entity.EventTypeConnectionConnecting && event.SessionID == sessionID {
					return true
				}
			}

			return false
		},
		gen.Identifier(),
	))

	properties.TestingRun(t)
}

// Feature: session-connection-flow, Property 8: Failed Connection Publishes Error Event
// *For any* connection failure in the Go service, a 'connection.failed' event SHALL be
// published with error_code and error_message fields populated.
// **Validates: Requirements 3.1**

func TestFailedEventPublication_Property8(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 8.1: Connection failure publishes connection.failed event
	properties.Property("connection failure publishes failed event", prop.ForAll(
		func(sessionID string) bool {
			if sessionID == "" {
				return true // skip empty inputs
			}

			ctx := context.Background()
			db := setupTestDB(t)
			repo := persistence.NewSessionRepository(db)
			waClient := NewWhatsAppClientMock()
			publisher := NewEventPublisherMock()

			// Create session
			session := entity.NewSession(sessionID, "Test Session")
			_ = repo.Create(ctx, session)

			// Make connection fail
			waClient.ConnectFn = func(ctx context.Context, sid string) error {
				return errors.ErrConnectionFailed.WithMessage("simulated failure")
			}

			uc := usecase.NewSessionUseCase(repo, waClient, publisher, nil)

			// Call reconnect (will fail)
			_ = uc.ReconnectSessionWithJID(ctx, sessionID, "")

			// Verify failed event was published
			for _, event := range publisher.GetEvents() {
				if event.Type == entity.EventTypeConnectionFailed && event.SessionID == sessionID {
					return true
				}
			}

			return false
		},
		gen.Identifier(),
	))

	// Property 8.2: Failed event contains error_code field
	properties.Property("failed event contains error_code", prop.ForAll(
		func(sessionID string) bool {
			if sessionID == "" {
				return true // skip empty inputs
			}

			ctx := context.Background()
			db := setupTestDB(t)
			repo := persistence.NewSessionRepository(db)
			waClient := NewWhatsAppClientMock()
			publisher := NewEventPublisherMock()

			// Create session
			session := entity.NewSession(sessionID, "Test Session")
			_ = repo.Create(ctx, session)

			// Make connection fail
			waClient.ConnectFn = func(ctx context.Context, sid string) error {
				return errors.ErrConnectionFailed.WithMessage("simulated failure")
			}

			uc := usecase.NewSessionUseCase(repo, waClient, publisher, nil)

			// Call reconnect (will fail)
			_ = uc.ReconnectSessionWithJID(ctx, sessionID, "")

			// Find failed event and check error_code
			for _, event := range publisher.GetEvents() {
				if event.Type == entity.EventTypeConnectionFailed {
					var data entity.ConnectionFailedData
					if err := event.UnmarshalData(&data); err != nil {
						return false
					}
					return data.ErrorCode != ""
				}
			}

			return false
		},
		gen.Identifier(),
	))

	// Property 8.3: Failed event contains error_message field
	properties.Property("failed event contains error_message", prop.ForAll(
		func(sessionID string) bool {
			if sessionID == "" {
				return true // skip empty inputs
			}

			ctx := context.Background()
			db := setupTestDB(t)
			repo := persistence.NewSessionRepository(db)
			waClient := NewWhatsAppClientMock()
			publisher := NewEventPublisherMock()

			// Create session
			session := entity.NewSession(sessionID, "Test Session")
			_ = repo.Create(ctx, session)

			// Make connection fail
			waClient.ConnectFn = func(ctx context.Context, sid string) error {
				return errors.ErrConnectionFailed.WithMessage("simulated failure")
			}

			uc := usecase.NewSessionUseCase(repo, waClient, publisher, nil)

			// Call reconnect (will fail)
			_ = uc.ReconnectSessionWithJID(ctx, sessionID, "")

			// Find failed event and check error_message
			for _, event := range publisher.GetEvents() {
				if event.Type == entity.EventTypeConnectionFailed {
					var data entity.ConnectionFailedData
					if err := event.UnmarshalData(&data); err != nil {
						return false
					}
					return data.ErrorMessage != ""
				}
			}

			return false
		},
		gen.Identifier(),
	))

	// Property 8.4: Failed event has correct session_id
	properties.Property("failed event has correct session_id", prop.ForAll(
		func(sessionID string) bool {
			if sessionID == "" {
				return true // skip empty inputs
			}

			ctx := context.Background()
			db := setupTestDB(t)
			repo := persistence.NewSessionRepository(db)
			waClient := NewWhatsAppClientMock()
			publisher := NewEventPublisherMock()

			// Create session
			session := entity.NewSession(sessionID, "Test Session")
			_ = repo.Create(ctx, session)

			// Make connection fail
			waClient.ConnectFn = func(ctx context.Context, sid string) error {
				return errors.ErrConnectionFailed.WithMessage("simulated failure")
			}

			uc := usecase.NewSessionUseCase(repo, waClient, publisher, nil)

			// Call reconnect (will fail)
			_ = uc.ReconnectSessionWithJID(ctx, sessionID, "")

			// Find failed event and check session_id
			for _, event := range publisher.GetEvents() {
				if event.Type == entity.EventTypeConnectionFailed {
					return event.SessionID == sessionID
				}
			}

			return false
		},
		gen.Identifier(),
	))

	// Property 8.5: Failed event has valid timestamp
	properties.Property("failed event has valid timestamp", prop.ForAll(
		func(sessionID string) bool {
			if sessionID == "" {
				return true // skip empty inputs
			}

			ctx := context.Background()
			db := setupTestDB(t)
			repo := persistence.NewSessionRepository(db)
			waClient := NewWhatsAppClientMock()
			publisher := NewEventPublisherMock()

			// Create session
			session := entity.NewSession(sessionID, "Test Session")
			_ = repo.Create(ctx, session)

			// Make connection fail
			waClient.ConnectFn = func(ctx context.Context, sid string) error {
				return errors.ErrConnectionFailed.WithMessage("simulated failure")
			}

			uc := usecase.NewSessionUseCase(repo, waClient, publisher, nil)

			// Call reconnect (will fail)
			_ = uc.ReconnectSessionWithJID(ctx, sessionID, "")

			// Find failed event and check timestamp
			for _, event := range publisher.GetEvents() {
				if event.Type == entity.EventTypeConnectionFailed {
					return !event.Timestamp.IsZero()
				}
			}

			return false
		},
		gen.Identifier(),
	))

	// Property 8.6: Client unavailable publishes failed event with CLIENT_UNAVAILABLE code
	properties.Property("client unavailable publishes failed event with correct code", prop.ForAll(
		func(sessionID string) bool {
			if sessionID == "" {
				return true // skip empty inputs
			}

			ctx := context.Background()
			db := setupTestDB(t)
			repo := persistence.NewSessionRepository(db)
			publisher := NewEventPublisherMock()

			// Create session
			session := entity.NewSession(sessionID, "Test Session")
			_ = repo.Create(ctx, session)

			// Create usecase with nil waClient
			uc := usecase.NewSessionUseCase(repo, nil, publisher, nil)

			// Call reconnect (will fail due to nil client)
			_ = uc.ReconnectSessionWithJID(ctx, sessionID, "")

			// Find failed event and check error_code
			for _, event := range publisher.GetEvents() {
				if event.Type == entity.EventTypeConnectionFailed {
					var data entity.ConnectionFailedData
					if err := event.UnmarshalData(&data); err != nil {
						return false
					}
					return data.ErrorCode == "CLIENT_UNAVAILABLE"
				}
			}

			return false
		},
		gen.Identifier(),
	))

	properties.TestingRun(t)
}
