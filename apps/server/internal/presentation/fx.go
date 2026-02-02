package presentation

import (
	"time"

	"whatspire/internal/application/usecase"
	"whatspire/internal/infrastructure/config"
	"whatspire/internal/infrastructure/ratelimit"
	infraWs "whatspire/internal/infrastructure/websocket"
	"whatspire/internal/presentation/http"
	"whatspire/internal/presentation/ws"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// Module provides all presentation layer dependencies
var Module = fx.Module("presentation",
	fx.Provide(
		NewHTTPHandler,
		NewRouter,
		NewQRHandler,
		NewEventHandler,
	),
)

// NewHTTPHandler creates a new HTTP handler with health use case
func NewHTTPHandler(
	sessionUC *usecase.SessionUseCase,
	messageUC *usecase.MessageUseCase,
	healthUC *usecase.HealthUseCase,
	groupsUC *usecase.GroupsUseCase,
	reactionUC *usecase.ReactionUseCase,
	receiptUC *usecase.ReceiptUseCase,
	presenceUC *usecase.PresenceUseCase,
) *http.Handler {
	return http.NewHandler(sessionUC, messageUC, healthUC, groupsUC, reactionUC, receiptUC, presenceUC)
}

// NewRouter creates a new Gin router with all routes configured
func NewRouter(
	handler *http.Handler,
	cfg *config.Config,
) *gin.Engine {
	// Create rate limiter if enabled
	var rateLimiter *ratelimit.Limiter
	if cfg.RateLimit.Enabled {
		rateLimiterConfig := ratelimit.Config{
			Enabled:           cfg.RateLimit.Enabled,
			RequestsPerSecond: cfg.RateLimit.RequestsPerSecond,
			BurstSize:         cfg.RateLimit.BurstSize,
			ByIP:              cfg.RateLimit.ByIP,
			ByAPIKey:          cfg.RateLimit.ByAPIKey,
			CleanupInterval:   cfg.RateLimit.CleanupInterval,
			MaxAge:            cfg.RateLimit.MaxAge,
		}
		rateLimiter = ratelimit.NewLimiter(rateLimiterConfig)
	}

	routerConfig := http.RouterConfig{
		Debug:                cfg.Log.Level == "debug",
		EnableRequestLogging: cfg.Log.Level == "debug",
		RateLimiter:          rateLimiter,
		CORSConfig:           &cfg.CORS,
		APIKeyConfig:         &cfg.APIKey,
	}

	return http.NewRouter(handler, routerConfig)
}

// NewQRHandler creates a new QR WebSocket handler
func NewQRHandler(sessionUC *usecase.SessionUseCase, cfg *config.Config) *ws.QRHandler {
	qrConfig := ws.QRHandlerConfig{
		AuthTimeout:    cfg.WhatsApp.QRTimeout,
		WriteTimeout:   10 * time.Second,
		PingInterval:   30 * time.Second,
		AllowedOrigins: cfg.CORS.AllowedOrigins,
	}

	return ws.NewQRHandler(sessionUC, qrConfig)
}

// NewEventHandler creates a new Event WebSocket handler
func NewEventHandler(hub *infraWs.EventHub, cfg *config.Config) *ws.EventHandler {
	eventConfig := ws.EventHandlerConfig{
		PingInterval:   cfg.WebSocket.PingInterval,
		WriteTimeout:   cfg.WebSocket.PongTimeout,
		AuthTimeout:    10 * time.Second,
		AllowedOrigins: cfg.CORS.AllowedOrigins,
	}

	return ws.NewEventHandler(hub, eventConfig)
}
