package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"whatspire/internal/app"
	"whatspire/internal/infrastructure/config"
	"whatspire/internal/presentation/ws"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// Version is the application version
const Version = "2.0.0"

func main() {
	log.Printf("🚀 Starting Whatspire WhatsApp Service v%s", Version)
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
	)

	// Start the application
	startCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := fxApp.Start(startCtx); err != nil {
		log.Fatalf("❌ Failed to start application: %v", err)
	}

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("🛑 Received signal: %v - initiating graceful shutdown...", sig)

	// Stop the application gracefully
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer stopCancel()

	if err := fxApp.Stop(stopCtx); err != nil {
		log.Fatalf("❌ Failed to stop application gracefully: %v", err)
	}

	log.Println("✅ Application stopped gracefully")
}

// startServer starts the HTTP server with graceful shutdown
func startServer(
	lc fx.Lifecycle,
	router *gin.Engine,
	qrHandler *ws.QRHandler,
	eventHandler *ws.EventHandler,
	cfg *config.Config,
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
			log.Printf("🚀 WhatsApp service starting on %s", cfg.Server.Address())
			log.Printf("📁 Whatsmeow database: %s", cfg.WhatsApp.DBPath)
			log.Printf("🔌 WebSocket API URL: %s", cfg.WebSocket.URL)

			// Start server in a goroutine
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Printf("❌ Server error: %v", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("🛑 Shutting down HTTP server...")

			// Create a deadline for graceful shutdown
			shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			if err := srv.Shutdown(shutdownCtx); err != nil {
				log.Printf("⚠️  HTTP server shutdown error: %v", err)
				return fmt.Errorf("server shutdown error: %w", err)
			}

			log.Println("✅ HTTP server stopped gracefully")
			return nil
		},
	})
}
