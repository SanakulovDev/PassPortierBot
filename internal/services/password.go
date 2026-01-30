package services

import (
	"fmt"
	"log"
	"time"

	"passportier-bot/internal/vault"

	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
)

// SavePassword encrypts and saves credential to database.
func SavePassword(db *gorm.DB, userID int64, service, data string) error {
	userKey, ok := vault.GetKey(userID)
	if !ok {
		return fmt.Errorf("session not found")
	}

	return vault.UpsertCredential(db, userID, service, []byte(data), userKey)
}

// GetPassword retrieves and decrypts credential from database.
func GetPassword(db *gorm.DB, userID int64, service string) (string, error) {
	userKey, ok := vault.GetKey(userID)
	if !ok {
		return "", fmt.Errorf("session not found")
	}

	return vault.RetrieveCredential(db, userID, service, userKey)
}

// ScheduleExpiration edits message to show expiration notice after 10 seconds.
func ScheduleExpiration(b *telebot.Bot, msg *telebot.Message) {
	go func(m *telebot.Message) {
		time.Sleep(10 * time.Second)
		expiredText := "⏰ *Muddati tugadi*\n\n_Xavfsizlik sababli bu ma'lumot yashirildi._"
		if _, err := b.Edit(m, expiredText, telebot.ModeMarkdown); err != nil {
			log.Printf("Warning: Failed to edit expired message (id %d): %v", m.ID, err)
		}
	}(msg)
}
