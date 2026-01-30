package handlers

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"passportier-bot/internal/services"

	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
)

// HandleText returns the text handler for hash-based save/retrieve (#service).
func HandleText(b *telebot.Bot, db *gorm.DB) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		// Delete user message for security
		defer func() {
			if err := b.Delete(c.Message()); err != nil {
				log.Println("Warning: Failed to delete user message:", err)
			}
		}()

		text := strings.ToLower(strings.TrimSpace(c.Text()))

		if !strings.HasPrefix(text, "#") {
			return c.Send("⚠️ Ma'lumot saqlash uchun `#xizmat_nomi ma'lumot` ko'rinishida yozing.\nOlish uchun esa shunchaki `#xizmat_nomi` deb yozing.")
		}

		serviceName, data := parseHashInput(text)
		if serviceName == "" {
			return c.Send("⚠️ Iltimos, xizmat nomini hash bilan yozing. Misol: `#instagram` yoki `#instagram parol123`")
		}

		if data != "" {
			return handleSave(c, b, db, serviceName, data)
		}
		return handleRetrieve(c, b, db, serviceName)
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

// handleSave saves password with loading indicator.
func handleSave(c telebot.Context, b *telebot.Bot, db *gorm.DB, serviceName, data string) error {
	msg, _ := b.Send(c.Sender(), "⏳ Saqlanmoqda...")

	if err := services.SavePassword(db, c.Sender().ID, serviceName, data); err != nil {
		b.Delete(msg)
		log.Printf("Save error: %v", err)
		return c.Send("❌ Saqlash xatosi. Sessiya yopiq bo'lishi mumkin.")
	}

	b.Delete(msg)
	return c.Send(fmt.Sprintf("✅ *%s* saqlandi!", serviceName), telebot.ModeMarkdown)
}

// handleRetrieve retrieves password with expiration notice.
func handleRetrieve(c telebot.Context, b *telebot.Bot, db *gorm.DB, serviceName string) error {
	decrypted, err := services.GetPassword(db, c.Sender().ID, serviceName)
	if err != nil {
		log.Printf("[ERROR] Retrieve failed: %v", err)
		return c.Send(fmt.Sprintf("❌ *%s* bo'yicha ma'lumot topilmadi.", serviceName), telebot.ModeMarkdown)
	}

	msgText := fmt.Sprintf("🔑 *%s*\n\n`%s`\n\n⚠️ _Bu xabar xavfsizlik uchun 10 soniyadan so'ng yashiriladi._", serviceName, decrypted)
	sentMsg, err := b.Send(c.Sender(), msgText, telebot.ModeMarkdown)
	if err != nil {
		return err
	}

	services.ScheduleExpiration(b, sentMsg)
	return nil
}
