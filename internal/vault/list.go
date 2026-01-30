package vault

import (
	"passportier-bot/internal/models"

	"gorm.io/gorm"
)

// ListEntries returns all password entries for a user (without decryption).
func ListEntries(db *gorm.DB, userID int64) ([]models.PasswordEntry, error) {
	var entries []models.PasswordEntry
	result := db.Where("user_id = ?", userID).Order("service ASC").Find(&entries)
	return entries, result.Error
}
