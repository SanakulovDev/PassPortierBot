package vault

import (
	"passportier-bot/internal/crypto"
	"passportier-bot/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UpsertCredential encrypts the data and upserts it into the database.
// Uses PostgreSQL's ON CONFLICT clause for atomic insert-or-update.
func UpsertCredential(db *gorm.DB, userID int64, service string, plainData []byte, key []byte) error {
	encrypted, err := crypto.Encrypt(plainData, key)
	if err != nil {
		return err
	}

	entry := buildEntry(userID, service, encrypted)

	// Upsert: conflict on (user_id, service) -> update encrypted_data
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "service"}},
		DoUpdates: clause.AssignmentColumns([]string{"encrypted_data", "updated_at"}),
	}).Create(&entry).Error
}

// buildEntry constructs a PasswordEntry model from the given parameters.
func buildEntry(userID int64, service string, encrypted []byte) models.PasswordEntry {
	return models.PasswordEntry{
		UserID:        userID,
		Service:       service,
		EncryptedData: encrypted,
	}
}
