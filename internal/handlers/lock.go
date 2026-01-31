package handlers

import (
	"context"
	"log"

	"passportier-bot/internal/security"

	"gopkg.in/telebot.v3"
)

// HandleLock returns the /lock command handler for manual session termination.
// This allows users to instantly close their session for security.
func HandleLock(b *telebot.Bot, sm *security.SessionManager) telebot.HandlerFunc {
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
		ctx := context.Background()

		// Check if session exists before deletion
		_, err := sm.GetSession(ctx, userID)
		existed := err == nil

		// Terminate session (idempotent)
		sm.ClearSession(ctx, userID)

		if existed {
			log.Printf("[SESSION] User %d manually locked session", userID)
			return c.Send("🔒 *Sessiya yopildi.*\n\nSizning seyfingiz qulflandi. Qayta ochish uchun `/unlock` buyrug'ini yuboring.", telebot.ModeMarkdown)
		}

		return c.Send("ℹ️ Faol sessiya topilmadi. Sessiyani ochish uchun `/unlock` buyrug'ini yuboring.", telebot.ModeMarkdown)
	}
}
