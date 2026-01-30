package services

import (
	"log"

	"passportier-bot/internal/vault"
)

// UnlockSession stores the user's passphrase in RAM for the session.
// Zero-Knowledge: The passphrase is used directly for per-encryption key derivation.
// No salt is stored in DB; each encryption generates its own unique salt.
func UnlockSession(telegramID int64, passphrase string) {
	vault.SetKey(telegramID, passphrase)
	log.Printf("[DEBUG] User %d unlocked session.", telegramID)
}
