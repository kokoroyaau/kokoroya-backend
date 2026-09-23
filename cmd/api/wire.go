//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"

	"kokoroya-backend/config"
	"kokoroya-backend/internal/app"
	"kokoroya-backend/internal/database"
	"kokoroya-backend/internal/email"
	"kokoroya-backend/pkg/logger"
)

func InitializeApp() (*app.App, error) {
	wire.Build(
		config.ProviderSet,
		logger.ProviderSet,
		database.ProviderSet,
		email.ProviderSet,
		app.ProviderSet,
	)
	return nil, nil
}
