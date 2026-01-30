package vault

import (
	"strings"

	"passportier-bot/internal/crypto"
	"passportier-bot/internal/models"

	"gorm.io/gorm"
)

// RetrieveCredential finds and decrypts the credential for the given service.
// Returns the decrypted plaintext or crypto.ErrInvalidPassword if key is wrong.
func RetrieveCredential(db *gorm.DB, userID int64, service string, userKey string) (string, error) {
	entry, err := findEntry(db, userID, service)
	if err != nil {
		return "", err
	}

	cm := crypto.NewCryptoManager()
	plaintext, err := cm.Decrypt(entry.EncryptedData, userKey)
	if err != nil {
		// Zero-Knowledge: wrong password manifests as decryption failure
		return "", err
	}

	return plaintext, nil
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
