package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"whatspire/internal/app"
	"whatspire/internal/infrastructure/config"
	"whatspire/internal/infrastructure/logger"
	"whatspire/internal/presentation/ws"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// Version is the application version
const Version = "2.0.0"

func main() {
	// Create logger for main
	log := logger.New("info", "text")

	log.Infof("Starting Whatspire WhatsApp Service v%s", Version)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	fxApp := fx.New(
		// Include all application modules
		app.Module,

		// Invoke the server startup
		fx.Invoke(startServer),

		// Configure graceful shutdown timeout
		fx.StopTimeout(45*time.Second), // Allow 45 seconds for graceful shutdown

		// Suppress Fx verbose logging
		fx.NopLogger,
	)

	// Start the application
	startCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := fxApp.Start(startCtx); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	// Wait for shutdown signal
	sig := <-sigChan
	log.Infof("Received shutdown signal: %v - initiating graceful shutdown", sig)

	// Stop the application gracefully
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer stopCancel()

	if err := fxApp.Stop(stopCtx); err != nil {
		log.Fatalf("Failed to stop application gracefully: %v", err)
	}

	log.Info("Application stopped gracefully")
}

// startServer starts the HTTP server with graceful shutdown
func startServer(
	lc fx.Lifecycle,
	router *gin.Engine,
	qrHandler *ws.QRHandler,
	eventHandler *ws.EventHandler,
	cfg *config.Config,
	log *logger.Logger,
) {
	// Register QR WebSocket routes on the router
	qrHandler.RegisterRoutes(router)

	// Register Event WebSocket routes on the router
	eventHandler.RegisterRoutes(router)

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.Address(),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.WithFields(map[string]interface{}{
				"address":       cfg.Server.Address(),
				"db_path":       cfg.WhatsApp.DBPath,
				"websocket_url": cfg.WebSocket.URL,
			}).Info("WhatsApp service starting")

			// Start server in a goroutine
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.WithError(err).Error("HTTP server encountered an error")
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Initiating HTTP server graceful shutdown")

			// Create a deadline for graceful shutdown
			shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			if err := srv.Shutdown(shutdownCtx); err != nil {
				log.WithError(err).Warn("HTTP server shutdown encountered an error")
				return fmt.Errorf("server shutdown error: %w", err)
			}

			log.Info("HTTP server stopped gracefully")
			return nil
		},
	})
}
