package vault

import (
	"sync"
	"time"
)

// Session represents a user's active session.
// Zero-Knowledge: We store the passphrase, NOT a derived key.
// The key is derived fresh for each encrypt/decrypt operation with unique salt.
type Session struct {
	passphrase string      // User's passphrase (never stored to disk)
	version    int64       // Incremented on every new unlock to invalidate old checks
	timer      *time.Timer // The timer that will delete this session
}

var (
	// sessions stores active user sessions.
	// Key: User ID (int64)
	// Value: *Session
	sessions = make(map[int64]*Session)
	mu       sync.RWMutex
)

// SetKey securely stores the user's passphrase in RAM.
// It handles race conditions by cancelling old timers and using versioning.
func SetKey(userID int64, passphrase string) {
	mu.Lock()
	defer mu.Unlock()

	// 1. Cleanup existing session if any
	if oldSession, exists := sessions[userID]; exists {
		if oldSession.timer != nil {
			oldSession.timer.Stop()
		}
	}

	// 2. Create new session
	newSession := &Session{
		passphrase: passphrase,
		version:    time.Now().UnixNano(),
	}

	// 3. Setup auto-delete timer (30 minutes TTL)
	newSession.timer = time.AfterFunc(30*time.Minute, func() {
		cleanupSession(userID, newSession.version)
	})

	sessions[userID] = newSession
}

// GetKey retrieves the user's passphrase.
// Returns (passphrase, true) if found, ("", false) otherwise.
func GetKey(userID int64) (string, bool) {
	mu.RLock()
	defer mu.RUnlock()

	session, exists := sessions[userID]
	if !exists {
		return "", false
	}

	return session.passphrase, true
}

// DeleteKey manually removes a user's session.
func DeleteKey(userID int64) {
	mu.Lock()
	defer mu.Unlock()

	if session, exists := sessions[userID]; exists {
		if session.timer != nil {
			session.timer.Stop()
		}
		delete(sessions, userID)
	}
}

