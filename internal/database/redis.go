package database

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"kokoroya-backend/config"
)

// NewRedisClient creates a new Redis client using the given config.
func NewRedisClient(cfg *config.Config, log *logrus.Logger) (*redis.Client, error) {
	rc := cfg.Redis
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", rc.Host, rc.Port),
		Password: rc.Password,
		DB:       rc.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	log.Info("connected to redis")
	return client, nil
}
