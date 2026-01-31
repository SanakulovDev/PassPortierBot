package models

import "gorm.io/gorm"

// User model stores the per-user salt required for Argon2id key derivation.
// We do NOT store paswords or hashes here, only the Salt.
type User struct {
	gorm.Model
	TelegramID int64  `gorm:"uniqueIndex;not null"`
	Salt       []byte `gorm:"not null"` // Random salt for this user
	SessionTTL int64  `gorm:"default:1800"` // Session TTL in seconds (default 30 mins)
}
