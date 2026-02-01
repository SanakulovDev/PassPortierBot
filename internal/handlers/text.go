package handlers

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"passportier-bot/internal/security"
	"passportier-bot/internal/services"

	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
)

// HandleText returns the text handler for hash-based retrieval (#service).
// Saving via text (#service data) is deprecated in V2.0.
func HandleText(b *telebot.Bot, db *gorm.DB, sm *security.SessionManager) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		// Delete user message for security
		defer func() {
			if err := b.Delete(c.Message()); err != nil {
				log.Println("Warning: Failed to delete user message:", err)
			}
		}()

		text := strings.ToLower(strings.TrimSpace(c.Text()))

		if !strings.HasPrefix(text, "#") {
			return nil // Ignore non-command text
		}

		serviceName, data := parseHashInput(text)
		if serviceName == "" {
			return c.Send("⚠️ Xizmat nomini hash bilan yozing. Misol: `#instagram`")
		}

		// Block text-based saving - show WebApp button
		if data != "" {
			webAppURL := os.Getenv("WEBAPP_URL")
			if webAppURL == "" {
				webAppURL = "https://your-domain.com/add_password.html"
			}

			menu := &telebot.ReplyMarkup{ResizeKeyboard: true}
			btnWebApp := menu.WebApp("➕ Parol Qo'shish", &telebot.WebApp{
				URL: webAppURL,
			})
			menu.Reply(menu.Row(btnWebApp))

			return c.Send("🛑 Matn orqali saqlash o'chirilgan.\nQuyidagi tugmadan foydalaning:", menu)
		}

		return handleRetrieve(c, b, db, sm, serviceName)
	}
}

// parseHashInput parses #service data format.
func parseHashInput(text string) (service, data string) {
	cleanText := strings.TrimPrefix(text, "#")
	re := regexp.MustCompile(`\s+`)
	parts := re.Split(strings.TrimSpace(cleanText), 2)

	service = strings.TrimSpace(parts[0])
	if len(parts) > 1 {
		data = strings.TrimSpace(parts[1])
	}
	return
}



// handleRetrieve retrieves password with countdown timer.
func handleRetrieve(c telebot.Context, b *telebot.Bot, db *gorm.DB, sm *security.SessionManager, serviceName string) error {
	decrypted, err := services.GetPassword(context.Background(), db, sm, c.Sender().ID, serviceName)
	if err != nil {
		log.Printf("[ERROR] Retrieve failed: %v", err)
		return c.Send(fmt.Sprintf("❌ *%s* bo'yicha ma'lumot topilmadi yoki sessiya yopiq.", serviceName), telebot.ModeMarkdown)
	}

	// Original text without countdown (for countdown updates)
	originalText := fmt.Sprintf("🔑 *%s*\n\n`%s`", serviceName, decrypted)
	msgText := fmt.Sprintf("%s\n\n⏱ _Yashirilishiga 30 soniya qoldi..._", originalText)

	sentMsg, err := b.Send(c.Sender(), msgText, telebot.ModeMarkdown)
	if err != nil {
		return err
	}

	services.ScheduleCountdown(b, sentMsg, originalText)
	return nil
}
