package vault

import "log"

// cleanupSession is called by the timer. It safely removes the session
// ONLY if the versions match (handling the race condition).
func cleanupSession(userID int64, version int64) {
	mu.Lock()
	defer mu.Unlock()

	session, exists := sessions[userID]
	if !exists {
		return // Already deleted
	}

	// RACE CONDITION FIX:
	// Only delete if the session version matches the one that set the timer.
	// If the user re-logged in (SetKey called again), the versions won't match,
	// and we should NOT delete the new active session.
	if session.version == version {
		log.Printf("[VAULT] Auto-expiring session for user %d", userID)
		delete(sessions, userID)
	}
}
