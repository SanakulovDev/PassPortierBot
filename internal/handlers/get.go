package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"

	"passportier-bot/internal/security"
	"passportier-bot/internal/services"

	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
)

// HandleGet returns the /get command handler for password retrieval.
func HandleGet(b *telebot.Bot, db *gorm.DB, sm *security.SessionManager) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		// Delete message for security
		if err := b.Delete(c.Message()); err != nil {
			log.Println("Warning: Failed to delete get message:", err)
		}

		serviceName := parseServiceName(c)
		if serviceName == "" {
			return c.Send("⚠️ Qaysi xizmatni qidiryapsiz? Misol: /get google")
		}

		decrypted, err := services.GetPassword(context.Background(), db, sm, c.Sender().ID, serviceName)
		if err != nil {
			log.Printf("[ERROR] Get failed for User %d Service %s: %v", c.Sender().ID, serviceName, err)
			return c.Send("❌ Topilmadi yoki sessiya yopiq. `/unlock` ni tekshiring.", telebot.ModeMarkdown)
		}

		return c.Send(fmt.Sprintf("🔑 *%s*\n\n%s", serviceName, decrypted), telebot.ModeMarkdown)
	}
}

// parseServiceName extracts service name from /get command.
func parseServiceName(c telebot.Context) string {
	serviceName := strings.TrimSpace(c.Message().Payload)
	if serviceName == "" {
		args := strings.SplitN(c.Text(), " ", 2)
		if len(args) >= 2 {
			serviceName = strings.TrimSpace(args[1])
		}
	}
	return serviceName
}
