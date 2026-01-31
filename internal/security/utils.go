package security

import (
	"context"
	"fmt"
)

// fmtSessionKey formats the Redis key for a user session.
func fmtSessionKey(userID int64) string {
	return fmt.Sprintf("session:%d", userID)
}

// ClearSession removes the session key immediately.
func (sm *SessionManager) ClearSession(ctx context.Context, userID int64) error {
	redisKey := fmtSessionKey(userID)
	return sm.client.Del(ctx, redisKey).Err()
}
