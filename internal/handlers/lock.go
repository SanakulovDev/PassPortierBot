package handlers

import (
	"log"

	"passportier-bot/internal/vault"

	"gopkg.in/telebot.v3"
)

// HandleLock returns the /lock command handler for manual session termination.
// This allows users to instantly close their session for security.
func HandleLock(b *telebot.Bot) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		// Private chat only
		if c.Chat().Type != telebot.ChatPrivate {
			return nil
		}

		// Delete command message for security
		if err := b.Delete(c.Message()); err != nil {
			log.Println("Warning: Failed to delete lock message:", err)
		}

		userID := c.Sender().ID

		// Check if session exists before deletion
		_, existed := vault.GetKey(userID)

		// Terminate session (idempotent - succeeds even if no session)
		vault.DeleteKey(userID)

		if existed {
			log.Printf("[VAULT] User %d manually locked session", userID)
			return c.Send("🔒 *Sessiya yopildi.*\n\nSizning seyfingiz qulflandi. Qayta ochish uchun `/unlock` buyrug'ini yuboring.", telebot.ModeMarkdown)
		}

		return c.Send("ℹ️ Faol sessiya topilmadi. Sessiyani ochish uchun `/unlock` buyrug'ini yuboring.", telebot.ModeMarkdown)
	}
}
