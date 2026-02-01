package vault

import (
	"passportier-bot/internal/models"

	"gorm.io/gorm"
)

// GetEntry retrieves a single password entry by service name.
func GetEntry(db *gorm.DB, userID int64, service string) (*models.PasswordEntry, error) {
	var entry models.PasswordEntry
	err := db.Where("user_id = ? AND service = ?", userID, service).First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}
