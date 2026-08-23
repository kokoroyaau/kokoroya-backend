package session

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func sessionKey(jti string) string { return "session:" + jti }

type Manager struct {
	client *redis.Client
}

func NewManager(client *redis.Client) *Manager {
	return &Manager{client: client}
}

func (m *Manager) Create(ctx context.Context, userID int64, jti string, ttl time.Duration) error {
	return m.client.Set(ctx, sessionKey(jti), strconv.FormatInt(userID, 10), ttl).Err()
}

func (m *Manager) IsValid(ctx context.Context, jti string) (bool, error) {
	n, err := m.client.Exists(ctx, sessionKey(jti)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (m *Manager) Revoke(ctx context.Context, jti string) error {
	return m.client.Del(ctx, sessionKey(jti)).Err()
}
