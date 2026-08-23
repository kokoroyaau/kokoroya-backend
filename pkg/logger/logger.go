package logger

import (
	"os"

	"github.com/sirupsen/logrus"

	"kokoroya-backend/config"
)

// New creates a configured logrus logger instance from the app config.
func New(cfg *config.Config) *logrus.Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.JSONFormatter{})

	if cfg.App.Env == "development" {
		log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	return log
}
