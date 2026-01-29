package models

import "gorm.io/gorm"

type PasswordEntry struct {
	gorm.Model
	UserID        int64  `gorm:"index"`
	Service       string
	EncryptedData []byte // Shifrlangan login/parol
}