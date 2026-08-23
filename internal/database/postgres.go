package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"kokoroya-backend/config"
)

// NewPostgresConnection opens a connection to Postgres using the given config.
func NewPostgresConnection(cfg *config.Config, log *logrus.Logger) (*sql.DB, error) {
	pg := cfg.Postgres
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		pg.Host, pg.Port, pg.User, pg.Password, pg.DBName, pg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	log.Info("connected to postgres")
	return db, nil
}
