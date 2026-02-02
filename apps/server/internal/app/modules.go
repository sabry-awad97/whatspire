package app

import (
	"whatspire/internal/application"
	"whatspire/internal/infrastructure"
	"whatspire/internal/infrastructure/config"
	"whatspire/internal/presentation"

	"go.uber.org/fx"
)

// Module aggregates all application modules for easy import
var Module = fx.Options(
	config.Module,
	infrastructure.Module,
	application.Module,
	presentation.Module,
)
