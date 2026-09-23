package app

import (
	"database/sql"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"kokoroya-backend/config"
	"kokoroya-backend/internal/email"
)

type App struct {
	Config *config.Config
	Logger *logrus.Logger
	DB     *sql.DB
	Redis  *redis.Client
	Email  email.Service
}

func NewApp(cfg *config.Config, log *logrus.Logger, db *sql.DB, rdb *redis.Client, emailService email.Service) *App {
	return &App{
		Config: cfg,
		Logger: log,
		DB:     db,
		Redis:  rdb,
		Email:  emailService,
	}
}
