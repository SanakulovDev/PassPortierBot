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

	return vault.UpsertCredential(db, userID, service, data, userKey)
}

// GetPassword retrieves and decrypts credential from database.
func GetPassword(db *gorm.DB, userID int64, service string) (string, error) {
	userKey, ok := vault.GetKey(userID)
	if !ok {
		return "", fmt.Errorf("session not found")
	}

	return vault.RetrieveCredential(db, userID, service, userKey)
}

// ScheduleCountdown shows real-time countdown from 30 to 0 seconds.
// Updates message every 5 seconds to show remaining time.
func ScheduleCountdown(b *telebot.Bot, msg *telebot.Message, originalText string) {
	go func(m *telebot.Message, text string) {
		remaining := 30

		for remaining > 0 {
			time.Sleep(5 * time.Second)
			remaining -= 5

			if remaining > 0 {
				countdown := fmt.Sprintf("%s\n\n⏱ _Yashirilishiga %d soniya qoldi..._", text, remaining)
				if _, err := b.Edit(m, countdown, telebot.ModeMarkdown); err != nil {
					log.Printf("Countdown edit error: %v", err)
					return
				}
			}
		}

		// Final: hide the message
		expiredText := "⏰ *Muddati tugadi*\n\n_Xavfsizlik sababli yashirildi._"
		if _, err := b.Edit(m, expiredText, telebot.ModeMarkdown); err != nil {
			log.Printf("Expiration edit error: %v", err)
		}
	}(msg, originalText)
}
