package vault

import (
	"sync"
	"time"
)

// Session represents a user's active session.
type Session struct {
	key     []byte      // The AES key (32 bytes)
	version int64       // Incremented on every new unlock to invalidate old checks
	timer   *time.Timer // The timer that will delete this session
}

var (
	// sessions stores active user sessions.
	// Key: User ID (int64)
	// Value: *Session
	sessions = make(map[int64]*Session)
	mu       sync.RWMutex
)

// SetKey securely stores the user's AES key in RAM.
// It handles race conditions by cancelling old timers and using versioning.
func SetKey(userID int64, key []byte) {
	mu.Lock()
	defer mu.Unlock()

	// 1. Cleanup existing session if any (stop timer, zero memory)
	if oldSession, exists := sessions[userID]; exists {
		// Stop the old timer to prevent it from firing and deleting the new key
		if oldSession.timer != nil {
			oldSession.timer.Stop()
		}
		// Securely wipe the old key from memory immediately
		wipe(oldSession.key)
	}

	// 2. defensive copy: never store the slice passed by caller directly
	// to avoid aliasing issues.
	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)

	// 3. Create new session
	newSession := &Session{
		key:     keyCopy,
		version: time.Now().UnixNano(), // Unique version for this session
	}

	// 4. Setup auto-delete timer
	// We pass the version to the callback to ensure we only delete THIS specific session
	// if it's still active when the timer fires.
	newSession.timer = time.AfterFunc(30*time.Minute, func() {
		cleanupSession(userID, newSession.version)
	})

	sessions[userID] = newSession
}

// GetKey retrieves a defensive copy of the user's key.
// Returns (key, true) if found, (nil, false) otherwise.
func GetKey(userID int64) ([]byte, bool) {
	mu.RLock()
	defer mu.RUnlock()

	session, exists := sessions[userID]
	if !exists {
		return nil, false
	}

	// Return a copy so the caller cannot modify the vault's internal data
	keyCopy := make([]byte, len(session.key))
	copy(keyCopy, session.key)

	return keyCopy, true
}

// DeleteKey manually removes a user's session and wipes memory.
func DeleteKey(userID int64) {
	mu.Lock()
	defer mu.Unlock()

	if session, exists := sessions[userID]; exists {
		if session.timer != nil {
			session.timer.Stop()
		}
		wipe(session.key)
		delete(sessions, userID)
	}
}

