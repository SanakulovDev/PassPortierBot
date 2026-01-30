package handlers

import (
	"log"
	"strings"

	"passportier-bot/internal/services"

	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
)

// HandleUnlock returns the /unlock command handler for session authentication.
func HandleUnlock(b *telebot.Bot, db *gorm.DB) telebot.HandlerFunc {
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

		if err := services.UnlockSession(db, c.Sender().ID, passphrase); err != nil {
			return c.Send("❌ Tizim xatosi.")
		}

		return c.Send("🔓 Sessiya ochildi! Sizning kalitingiz Argon2id bilan shifrlanib, 30 daqiqa RAMda saqlanadi.")
	}
}

// parsePassphrase extracts passphrase from /unlock command.
func parsePassphrase(text string) string {
	args := strings.SplitN(text, " ", 2)
	if len(args) < 2 {
		return ""
	}
	return strings.TrimSpace(args[1])
}
