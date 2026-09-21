//go:build !wireinject
// +build !wireinject

package main

import (
	"kokoroya-backend/config"
	"kokoroya-backend/internal/app"
	"kokoroya-backend/internal/database"
	"kokoroya-backend/pkg/logger"
)

func InitializeApp() (*app.App, error) {
	configConfig, err := config.Load()
	if err != nil {
		return nil, err
	}
	logrusLogger := logger.New(configConfig)
	db, err := database.NewPostgresConnection(configConfig, logrusLogger)
	if err != nil {
		return nil, err
	}
	client, err := database.NewRedisClient(configConfig, logrusLogger)
	if err != nil {
		return nil, err
	}
	appApp := app.NewApp(configConfig, logrusLogger, db, client)
	return appApp, nil
}
