package handlers

import (
	"fmt"
	"log"
	"strings"
	"time"

	"passportier-bot/internal/crypto"
	"passportier-bot/internal/models"
	"passportier-bot/internal/vault"

	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
)

// HandleList returns the /list command handler to show all saved credentials.
func HandleList(b *telebot.Bot, db *gorm.DB) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		// Delete message for security
		if err := b.Delete(c.Message()); err != nil {
			log.Println("Warning: Failed to delete list message:", err)
		}

		userKey, ok := vault.GetKey(c.Sender().ID)
		if !ok {
			return c.Send("🔒 Sessiya yopiq. Avval `/unlock [so'z]` buyrug'ini yuboring.", telebot.ModeMarkdown)
		}

		entries, err := vault.ListEntries(db, c.Sender().ID)
		if err != nil || len(entries) == 0 {
			return c.Send("📭 Hech qanday saqlangan ma'lumot topilmadi.")
		}

		msgText := buildListMessage(entries, userKey)
		sentMsg, err := b.Send(c.Sender(), msgText, telebot.ModeMarkdown)
		if err != nil {
			return err
		}

		// Auto-hide after 10 seconds
		scheduleListExpiration(b, sentMsg)
		return nil
	}
}

// buildListMessage creates formatted list of all credentials.
func buildListMessage(entries []models.PasswordEntry, userKey string) string {
	cm := crypto.NewCryptoManager()
	var sb strings.Builder
	sb.WriteString("📋 *Barcha saqlangan ma'lumotlar:*\n\n")

	for i, entry := range entries {
		decrypted, err := cm.Decrypt(entry.EncryptedData, userKey)
		if err != nil {
			sb.WriteString(fmt.Sprintf("%d. *%s*: ❌ _shifr xatosi_\n", i+1, entry.Service))
		} else {
			sb.WriteString(fmt.Sprintf("%d. *%s*: `%s`\n", i+1, entry.Service, decrypted))
		}
	}

	sb.WriteString("\n⚠️ _Bu xabar 10 soniyadan so'ng yashiriladi._")
	return sb.String()
}

// scheduleListExpiration hides list message after 10 seconds.
func scheduleListExpiration(b *telebot.Bot, msg *telebot.Message) {
	go func(m *telebot.Message) {
		time.Sleep(10 * time.Second)
		expiredText := "⏰ *Muddati tugadi*\n\n_Xavfsizlik sababli bu ma'lumot yashirildi._"
		if _, err := b.Edit(m, expiredText, telebot.ModeMarkdown); err != nil {
			log.Printf("Warning: Failed to edit expired list (id %d): %v", m.ID, err)
		}
	}(msg)
}
