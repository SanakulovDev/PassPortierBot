package models

import "gorm.io/gorm"

type PasswordEntry struct {
	gorm.Model
	UserID        int64  `gorm:"index"`
	Service       string
	EncryptedData string // Base64 encoded: Salt + Nonce + Ciphertext
}