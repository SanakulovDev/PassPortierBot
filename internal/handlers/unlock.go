package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"passportier-bot/internal/models"
	"passportier-bot/internal/security"
	"passportier-bot/internal/services"

	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
)

// HandleUnlock returns the /unlock command handler for session authentication.
func HandleUnlock(b *telebot.Bot, sm *security.SessionManager, db *gorm.DB) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		// Private chat only
		if c.Chat().Type != telebot.ChatPrivate {
			return nil
		}

		// Delete message for security
		if err := b.Delete(c.Message()); err != nil {
			log.Println("Warning: Failed to delete unlock message:", err)
		}

		passphrase := parsePassphrase(c.Text())
		if passphrase == "" {
			return c.Send("⚠️ Iltimos, maxfiy so'z kiriting! Misol: `/unlock mySecretPass`", telebot.ModeMarkdown)
		}

		// Fetch user settings for TTL
		var user models.User
		if err := db.First(&user, "telegram_id = ?", c.Sender().ID).Error; err != nil {
			// If user not found, use default 30 mins
			user.SessionTTL = 1800
		}
		
		ttl := time.Duration(user.SessionTTL) * time.Second
		if ttl <= 0 {
			ttl = 30 * time.Minute
		}

		if err := services.UnlockSession(context.Background(), sm, c.Sender().ID, passphrase, ttl); err != nil {
			return c.Send("❌ Sessiyani ochishda xatolik yuz berdi.")
		}

		return c.Send(fmt.Sprintf("🔓 Sessiya ochildi! Kalitingiz %s davomida Redisda saqlanadi.", formatDuration(int64(ttl.Seconds()))))
	}
}

func formatDuration(seconds int64) string {
	if seconds < 60 {
		return fmt.Sprintf("%d soniya", seconds)
	} else if seconds < 3600 {
		return fmt.Sprintf("%d daqiqa", seconds/60)
	}
	return fmt.Sprintf("%d soat", seconds/3600)
}

// parsePassphrase extracts passphrase from /unlock command.
func parsePassphrase(text string) string {
	args := strings.SplitN(text, " ", 2)
	if len(args) < 2 {
		return ""
	}
	return strings.TrimSpace(args[1])
}
