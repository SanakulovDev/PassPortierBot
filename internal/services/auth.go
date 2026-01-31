package services

import (
	"context"
	"time"

	"passportier-bot/internal/security"
)

// UnlockSession stores the user's passphrase in Redis for the session.
// Zero-Knowledge: The passphrase is used directly for per-encryption key derivation.
// No salt is stored in DB; each encryption generates its own unique salt.
func UnlockSession(ctx context.Context, sm *security.SessionManager, userID int64, passphrase string, ttl time.Duration) error {
	return sm.SetSession(ctx, userID, passphrase, ttl)
}
