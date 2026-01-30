package services

import (
	"crypto/rand"
	"log"

	"passportier-bot/internal/crypto"
	"passportier-bot/internal/models"
	"passportier-bot/internal/vault"

	"gorm.io/gorm"
)

// UnlockSession handles user authentication: get/create user, derive key, set vault.
func UnlockSession(db *gorm.DB, telegramID int64, passphrase string) error {
	user, err := getOrCreateUser(db, telegramID)
	if err != nil {
		return err
	}

	// Derive key using Argon2id with user's unique salt
	key := crypto.DeriveKey(passphrase, user.Salt)

	// Store in vault (RAM only, 30 min TTL)
	vault.SetKey(telegramID, key)
	log.Printf("[DEBUG] User %d unlocked session. Key set (Argon2id).", telegramID)

	return nil
}

// getOrCreateUser finds existing user or creates new one with random salt.
func getOrCreateUser(db *gorm.DB, telegramID int64) (*models.User, error) {
	var user models.User

	if err := db.First(&user, "telegram_id = ?", telegramID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return createNewUser(db, telegramID)
		}
		log.Printf("DB error fetching user: %v", err)
		return nil, err
	}

	return &user, nil
}

// createNewUser generates salt and creates user record.
func createNewUser(db *gorm.DB, telegramID int64) (*models.User, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		log.Printf("Salt gen error: %v", err)
		return nil, err
	}

	user := models.User{
		TelegramID: telegramID,
		Salt:       salt,
	}

	if err := db.Create(&user).Error; err != nil {
		log.Printf("User creation error: %v", err)
		return nil, err
	}

	return &user, nil
}
