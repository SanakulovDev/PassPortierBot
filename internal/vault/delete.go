package vault

import (
	"passportier-bot/internal/models"

	"gorm.io/gorm"
)

// DeleteEntry removes a password entry by service name.
func DeleteEntry(db *gorm.DB, userID int64, service string) error {
	return db.Where("user_id = ? AND service = ?", userID, service).
		Delete(&models.PasswordEntry{}).Error
}
