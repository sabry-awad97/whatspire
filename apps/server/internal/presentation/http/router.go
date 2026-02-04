package http

import (
	"whatspire/internal/domain/repository"
	"whatspire/internal/infrastructure/config"
	"whatspire/internal/infrastructure/metrics"
	"whatspire/internal/infrastructure/ratelimit"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RouterConfig holds configuration for the router
type RouterConfig struct {
	// Debug enables debug mode
	Debug bool
	// EnableRequestLogging enables request body logging
	EnableRequestLogging bool
	// RateLimiter is the rate limiter instance (optional)
	RateLimiter *ratelimit.Limiter
	// CORSConfig is the CORS configuration (optional, uses defaults if nil)
	CORSConfig *config.CORSConfig
	// APIKeyConfig is the API key authentication configuration (optional)
	APIKeyConfig *config.APIKeyConfig
	// APIKeyRepository is the API key repository for database-backed keys (optional)
	APIKeyRepository repository.APIKeyRepository
	// Metrics is the metrics instance (optional)
	Metrics *metrics.Metrics
	// MetricsConfig is the metrics configuration (optional)
	MetricsConfig *config.MetricsConfig
	// AuditLogger is the audit logger instance (optional)
	AuditLogger repository.AuditLogger
}

// DefaultRouterConfig returns the default router configuration
func DefaultRouterConfig() RouterConfig {
	return RouterConfig{
		Debug:                false,
		EnableRequestLogging: false,
		RateLimiter:          nil,
		CORSConfig:           nil,
		APIKeyConfig:         nil,
		APIKeyRepository:     nil,
		Metrics:              nil,
		MetricsConfig:        nil,
		AuditLogger:          nil,
	}
}

// setupRouter creates and configures a Gin router with middleware
func setupRouter(routerConfig RouterConfig) *gin.Engine {
	if !routerConfig.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Apply global middleware
	router.Use(gin.Recovery())
	router.Use(ErrorHandlerMiddleware())
	router.Use(RequestIDMiddleware())
	router.Use(LoggingMiddleware())

	// Apply CORS middleware
	if routerConfig.CORSConfig != nil {
		router.Use(CORSMiddlewareWithConfig(*routerConfig.CORSConfig))
	} else {
		router.Use(CORSMiddlewareWithConfig(config.CORSConfig{}))
	}

	router.Use(ContentTypeMiddleware())

	// Apply metrics middleware if configured
	if routerConfig.Metrics != nil {
		router.Use(MetricsMiddleware(routerConfig.Metrics))
	}

	// Apply rate limiting if configured
	if routerConfig.RateLimiter != nil {
		router.Use(RateLimitMiddleware(routerConfig.RateLimiter))
	}

	// Optional request body logging
	if routerConfig.EnableRequestLogging {
		router.Use(RequestBodyLoggerMiddleware())
	}

	return router
}

// registerRoutes registers all routes on the router
func registerRoutes(router *gin.Engine, handler *Handler, routerConfig RouterConfig) {
	// Health check routes (no auth required)
	router.GET("/health", handler.Health)
	router.GET("/ready", handler.Ready)

	// Metrics endpoint (no auth required)
	if routerConfig.MetricsConfig != nil && routerConfig.MetricsConfig.Enabled {
		metricsPath := routerConfig.MetricsConfig.Path
		if metricsPath == "" {
			metricsPath = "/metrics"
		}
		router.GET(metricsPath, gin.WrapH(promhttp.Handler()))
	}

	// API routes
	api := router.Group("/api")

	// Apply API key authentication to API routes if configured
	if routerConfig.APIKeyConfig != nil && routerConfig.APIKeyConfig.Enabled {
		api.Use(APIKeyMiddleware(*routerConfig.APIKeyConfig, routerConfig.AuditLogger, routerConfig.APIKeyRepository))
	}

	// Internal routes (called by Node.js API for session lifecycle) - require admin role
	internal := api.Group("/internal")
	if routerConfig.APIKeyConfig != nil && routerConfig.APIKeyConfig.Enabled {
		internal.Use(RoleAuthorizationMiddleware(config.RoleAdmin, routerConfig.APIKeyConfig))
	}
	internal.POST("/sessions/register", handler.RegisterSession)
	internal.POST("/sessions/:id/unregister", handler.UnregisterSession)
	internal.POST("/sessions/:id/status", handler.UpdateSessionStatus)
	internal.POST("/sessions/:id/reconnect", handler.ReconnectSession)
	internal.POST("/sessions/:id/disconnect", handler.DisconnectSession)
	internal.POST("/sessions/:id/history-sync", handler.ConfigureHistorySync)

	// Session routes (groups sync) - require write role for sync, read for list
	sessions := api.Group("/sessions")
	if routerConfig.APIKeyConfig != nil && routerConfig.APIKeyConfig.Enabled {
		sessions.POST("", handler.CreateSession) // Public endpoint - no auth required in development
		sessions.GET("", RoleAuthorizationMiddleware(config.RoleRead, routerConfig.APIKeyConfig), handler.ListSessions)
		sessions.GET("/:id", RoleAuthorizationMiddleware(config.RoleRead, routerConfig.APIKeyConfig), handler.GetSession)
		sessions.DELETE("/:id", RoleAuthorizationMiddleware(config.RoleWrite, routerConfig.APIKeyConfig), handler.DeleteSession)
		sessions.POST("/:id/groups/sync", RoleAuthorizationMiddleware(config.RoleWrite, routerConfig.APIKeyConfig), handler.SyncGroups)
		sessions.GET("/:id/contacts", RoleAuthorizationMiddleware(config.RoleRead, routerConfig.APIKeyConfig), handler.ListContacts)
		sessions.GET("/:id/chats", RoleAuthorizationMiddleware(config.RoleRead, routerConfig.APIKeyConfig), handler.ListChats)
	} else {
		sessions.POST("", handler.CreateSession) // Public endpoint - no auth required in development
		sessions.GET("", handler.ListSessions)
		sessions.GET("/:id", handler.GetSession)
		sessions.DELETE("/:id", handler.DeleteSession)
		sessions.POST("/:id/groups/sync", handler.SyncGroups)
		sessions.GET("/:id/contacts", handler.ListContacts)
		sessions.GET("/:id/chats", handler.ListChats)
	}

	// Contact routes - require read role
	contacts := api.Group("/contacts")
	if routerConfig.APIKeyConfig != nil && routerConfig.APIKeyConfig.Enabled {
		contacts.GET("/check", RoleAuthorizationMiddleware(config.RoleRead, routerConfig.APIKeyConfig), handler.CheckPhoneNumber)
		contacts.GET("/:jid/profile", RoleAuthorizationMiddleware(config.RoleRead, routerConfig.APIKeyConfig), handler.GetUserProfile)
	} else {
		contacts.GET("/check", handler.CheckPhoneNumber)
		contacts.GET("/:jid/profile", handler.GetUserProfile)
	}

	// Message routes - require write role
	messages := api.Group("/messages")
	if routerConfig.APIKeyConfig != nil && routerConfig.APIKeyConfig.Enabled {
		messages.POST("", RoleAuthorizationMiddleware(config.RoleWrite, routerConfig.APIKeyConfig), handler.SendMessage)
		messages.POST("/:messageId/reactions", RoleAuthorizationMiddleware(config.RoleWrite, routerConfig.APIKeyConfig), handler.SendReaction)
		messages.DELETE("/:messageId/reactions", RoleAuthorizationMiddleware(config.RoleWrite, routerConfig.APIKeyConfig), handler.RemoveReaction)
		messages.POST("/receipts", RoleAuthorizationMiddleware(config.RoleWrite, routerConfig.APIKeyConfig), handler.SendReadReceipt)
	} else {
		messages.POST("", handler.SendMessage)
		messages.POST("/:messageId/reactions", handler.SendReaction)
		messages.DELETE("/:messageId/reactions", handler.RemoveReaction)
		messages.POST("/receipts", handler.SendReadReceipt)
	}

	// Presence routes - require write role
	if routerConfig.APIKeyConfig != nil && routerConfig.APIKeyConfig.Enabled {
		api.POST("/presence", RoleAuthorizationMiddleware(config.RoleWrite, routerConfig.APIKeyConfig), handler.SendPresence)
	} else {
		api.POST("/presence", handler.SendPresence)
	}

	// Event routes - require read role for query, admin role for replay
	events := api.Group("/events")
	if routerConfig.APIKeyConfig != nil && routerConfig.APIKeyConfig.Enabled {
		events.GET("", RoleAuthorizationMiddleware(config.RoleRead, routerConfig.APIKeyConfig), handler.QueryEvents)
		events.GET("/:id", RoleAuthorizationMiddleware(config.RoleRead, routerConfig.APIKeyConfig), handler.GetEventByID)
		events.POST("/replay", RoleAuthorizationMiddleware(config.RoleAdmin, routerConfig.APIKeyConfig), handler.ReplayEvents)
	} else {
		events.GET("", handler.QueryEvents)
		events.GET("/:id", handler.GetEventByID)
		events.POST("/replay", handler.ReplayEvents)
	}

	// API Key routes - require admin role
	apikeys := api.Group("/apikeys")
	if routerConfig.APIKeyConfig != nil && routerConfig.APIKeyConfig.Enabled {
		apikeys.POST("", RoleAuthorizationMiddleware(config.RoleAdmin, routerConfig.APIKeyConfig), handler.CreateAPIKey)
		apikeys.DELETE("/:id", RoleAuthorizationMiddleware(config.RoleAdmin, routerConfig.APIKeyConfig), handler.RevokeAPIKey)
	} else {
		apikeys.POST("", handler.CreateAPIKey)
		apikeys.DELETE("/:id", handler.RevokeAPIKey)
	}
}

// NewRouter creates a new Gin router with a pre-configured handler
func NewRouter(handler *Handler, routerConfig RouterConfig) *gin.Engine {
	router := setupRouter(routerConfig)
	registerRoutes(router, handler, routerConfig)
	return router
}
