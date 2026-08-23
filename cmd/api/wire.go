//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"

	"kokoroya-backend/config"
	"kokoroya-backend/internal/app"
	"kokoroya-backend/internal/database"
	"kokoroya-backend/pkg/logger"
)

// InitializeApp wires config, logger, and database connections into an App.
func InitializeApp() (*app.App, error) {
	wire.Build(
		config.ProviderSet,
		logger.ProviderSet,
		database.ProviderSet,
		app.ProviderSet,
	)
	return nil, nil
}
