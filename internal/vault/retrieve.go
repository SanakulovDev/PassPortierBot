package vault

import (
	"fmt"
	"strings"

	"passportier-bot/internal/crypto"
	"passportier-bot/internal/models"

	"gorm.io/gorm"
)

// RetrieveCredential finds and decrypts the credential for the given service.
// Returns the decrypted plaintext or an error if not found or decryption fails.
func RetrieveCredential(db *gorm.DB, userID int64, service string, key []byte) (string, error) {
	entry, err := findEntry(db, userID, service)
	if err != nil {
		return "", err
	}

	decrypted, err := crypto.Decrypt(entry.EncryptedData, key)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(decrypted), nil
}

// findEntry queries the database for a matching credential entry.
// Performs case-insensitive partial matching on the service name.
func findEntry(db *gorm.DB, userID int64, service string) (*models.PasswordEntry, error) {
	var entry models.PasswordEntry
	serviceLower := "%" + strings.ToLower(service) + "%"

	result := db.Where("user_id = ? AND LOWER(service) LIKE ?", userID, serviceLower).First(&entry)
	if result.Error != nil {
		return nil, result.Error
	}

	return &entry, nil
}
