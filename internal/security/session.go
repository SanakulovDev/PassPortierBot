package security

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionManager struct {
	client *redis.Client
}

func NewSessionManager(client *redis.Client) *SessionManager {
	return &SessionManager{client: client}
}

// SetSession stores the session key with the specified TTL.
// Zero-Knowledge: We store the passphrase in Redis (Ram) with strictly limited TTL.
func (sm *SessionManager) SetSession(ctx context.Context, userID int64, key string, ttl time.Duration) error {
	redisKey := fmtSessionKey(userID)
	return sm.client.Set(ctx, redisKey, key, ttl).Err()
}

// GetSession retrieves the session key if it exists.
func (sm *SessionManager) GetSession(ctx context.Context, userID int64) (string, error) {
	redisKey := fmtSessionKey(userID)
	return sm.client.Get(ctx, redisKey).Result()
}
