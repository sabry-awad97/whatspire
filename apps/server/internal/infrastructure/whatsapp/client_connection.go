package whatsapp

import (
	"context"
	"fmt"
	"time"

	"whatspire/internal/domain/errors"
	"whatspire/internal/domain/repository"

	"go.mau.fi/whatsmeow"
)

// Connect establishes a connection for the given session
func (c *WhatsmeowClient) Connect(ctx context.Context, sessionID string) error {
	// Use circuit breaker if enabled
	if c.circuitBreaker != nil {
		_, err := c.circuitBreaker.Execute(ctx, func() (any, error) {
			return nil, c.connectInternal(ctx, sessionID)
		})
		return err
	}
	return c.connectInternal(ctx, sessionID)
}

// connectInternal performs the actual connection logic
func (c *WhatsmeowClient) connectInternal(ctx context.Context, sessionID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already connected
	if client, exists := c.clients[sessionID]; exists && client.IsConnected() {
		return nil
	}

	// Get or create device store
	device, err := c.getOrCreateDevice(ctx, sessionID)
	if err != nil {
		return err
	}

	// Create client
	client := whatsmeow.NewClient(device, c.logger)

	// Register event handler
	client.AddEventHandler(func(evt interface{}) {
		c.handleEvent(sessionID, client, evt)
	})

	// Connect with retry
	err = c.connectWithRetry(ctx, client)
	if err != nil {
		return err
	}

	c.clients[sessionID] = client

	// Store JID mapping if available
	if client.Store.ID != nil {
		c.sessionToJID[sessionID] = client.Store.ID.User
		c.logger.Infof("connectInternal: stored JID mapping: sessionID=%s, jidUser=%s", sessionID, client.Store.ID.User)
	}

	return nil
}

// connectWithRetry implements exponential backoff retry for connections
func (c *WhatsmeowClient) connectWithRetry(ctx context.Context, client *whatsmeow.Client) error {
	retryPolicy := NewRetryPolicy(RetryConfig{
		MaxAttempts:  c.config.MaxReconnects,
		InitialDelay: c.config.ReconnectDelay,
		MaxDelay:     time.Duration(c.config.MaxReconnects) * time.Minute,
		Multiplier:   2.0,
		JitterFactor: 0.1,
	})

	err := retryPolicy.Execute(ctx, func() error {
		return client.Connect()
	})

	if err != nil {
		return errors.ErrConnectionFailed.WithCause(err).WithMessage(
			fmt.Sprintf("failed to connect after %d attempts", c.config.MaxReconnects+1))
	}
	return nil
}

// Disconnect closes the connection for the given session
func (c *WhatsmeowClient) Disconnect(ctx context.Context, sessionID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	client, exists := c.clients[sessionID]
	if !exists {
		return errors.ErrSessionNotFound
	}

	client.Disconnect()
	delete(c.clients, sessionID)
	return nil
}

// GetQRChannel returns a channel that receives QR code events for authentication
func (c *WhatsmeowClient) GetQRChannel(ctx context.Context, sessionID string) (<-chan repository.QREvent, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get or create device store
	device, err := c.getOrCreateDevice(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Create client
	client := whatsmeow.NewClient(device, c.logger)

	// Create QR event channel
	qrChan := make(chan repository.QREvent, 10)

	// Register event handler for this session
	client.AddEventHandler(func(evt interface{}) {
		c.handleEvent(sessionID, client, evt)
	})

	// Start QR authentication in goroutine
	go func() {
		defer close(qrChan)

		// Get QR channel from whatsmeow
		waQRChan, err := client.GetQRChannel(ctx)
		if err != nil {
			qrChan <- repository.QREvent{
				Type:    "error",
				Message: err.Error(),
			}
			return
		}

		// Set timeout
		timeout := time.NewTimer(c.config.QRTimeout)
		defer timeout.Stop()

		// Connect to start QR generation
		err = client.Connect()
		if err != nil {
			qrChan <- repository.QREvent{
				Type:    "error",
				Message: err.Error(),
			}
			return
		}

		for {
			select {
			case <-ctx.Done():
				client.Disconnect()
				return

			case <-timeout.C:
				qrChan <- repository.QREvent{
					Type:    "timeout",
					Message: "QR authentication timed out",
				}
				client.Disconnect()
				return

			case evt, ok := <-waQRChan:
				if !ok {
					return
				}

				switch evt.Event {
				case "code":
					qrChan <- repository.QREvent{
						Type: "qr",
						Data: evt.Code,
					}

				case "success":
					// Store client and JID mapping
					c.mu.Lock()
					c.clients[sessionID] = client
					if client.Store.ID != nil {
						c.sessionToJID[sessionID] = client.Store.ID.User
						c.logger.Infof("GetQRChannel: stored JID mapping: sessionID=%s, jidUser=%s", sessionID, client.Store.ID.User)
					}
					c.mu.Unlock()

					qrChan <- repository.QREvent{
						Type: "authenticated",
						Data: client.Store.ID.String(),
					}
					return

				case "timeout":
					qrChan <- repository.QREvent{
						Type:    "timeout",
						Message: "QR code expired",
					}
					// New QR will be generated automatically
				}
			}
		}
	}()

	return qrChan, nil
}

// AutoReconnect attempts to reconnect all sessions that have stored credentials
// Returns a map of session ID to error (nil if successful)
func (c *WhatsmeowClient) AutoReconnect(ctx context.Context) map[string]error {
	results := make(map[string]error)

	// Get all stored sessions
	sessionIDs, err := c.GetStoredSessions(ctx)
	if err != nil {
		c.logger.Errorf("AutoReconnect: failed to get stored sessions: %v", err)
		return results
	}

	c.logger.Infof("AutoReconnect: found %d stored sessions", len(sessionIDs))

	// Attempt to reconnect each session
	for _, sessionID := range sessionIDs {
		c.logger.Infof("AutoReconnect: attempting to reconnect session: %s", sessionID)

		err := c.Connect(ctx, sessionID)
		if err != nil {
			c.logger.Errorf("AutoReconnect: failed to reconnect session %s: %v", sessionID, err)
			results[sessionID] = err
		} else {
			c.logger.Infof("AutoReconnect: successfully reconnected session: %s", sessionID)
			results[sessionID] = nil
		}
	}

	return results
}
