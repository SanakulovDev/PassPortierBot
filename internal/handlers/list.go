package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"passportier-bot/internal/crypto"
	"passportier-bot/internal/models"
	"passportier-bot/internal/vault"

	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
)

const (
	itemsPerPage = 5 // Items per page for pagination
)

// HandleList returns the /list command handler with pagination support.
func HandleList(b *telebot.Bot, db *gorm.DB) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		if err := b.Delete(c.Message()); err != nil {
			log.Println("Warning: Failed to delete list message:", err)
		}

		return showListPage(b, c, db, 0)
	}
}

// RegisterListCallbacks registers pagination callback handlers.
func RegisterListCallbacks(b *telebot.Bot, db *gorm.DB) {
	b.Handle(&telebot.InlineButton{Unique: "list_page"}, func(c telebot.Context) error {
		page, _ := strconv.Atoi(c.Data())
		return showListPage(b, c, db, page)
	})

	b.Handle(&telebot.InlineButton{Unique: "list_refresh"}, func(c telebot.Context) error {
		return showListPage(b, c, db, 0)
	})
}

// showListPage displays a paginated list of secrets.
func showListPage(b *telebot.Bot, c telebot.Context, db *gorm.DB, page int) error {
	userKey, ok := vault.GetKey(c.Sender().ID)
	if !ok {
		return c.Send("🔒 Sessiya yopiq. `/unlock [so'z]` buyrug'ini yuboring.", telebot.ModeMarkdown)
	}

	entries, err := vault.ListEntries(db, c.Sender().ID)
	if err != nil || len(entries) == 0 {
		return c.Send("📭 Saqlangan ma'lumotlar yo'q.")
	}

	totalPages := (len(entries) + itemsPerPage - 1) / itemsPerPage
	if page >= totalPages {
		page = totalPages - 1
	}
	if page < 0 {
		page = 0
	}

	start := page * itemsPerPage
	end := start + itemsPerPage
	if end > len(entries) {
		end = len(entries)
	}

	pageEntries := entries[start:end]
	msgText, keyboard := buildPageContent(pageEntries, userKey, page, totalPages, start)

	opts := &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdown,
		ReplyMarkup: keyboard,
	}

	if c.Callback() != nil {
		_, err = b.Edit(c.Message(), msgText, opts)
		return err
	}

	sentMsg, err := b.Send(c.Sender(), msgText, opts)
	if err != nil {
		return err
	}

	scheduleListExpiration(b, sentMsg)
	return nil
}

// buildPageContent creates message and keyboard for current page.
func buildPageContent(entries []models.PasswordEntry, userKey string, page, totalPages, startIdx int) (string, *telebot.ReplyMarkup) {
	cm := crypto.NewCryptoManager()
	markup := &telebot.ReplyMarkup{}
	var rows []telebot.Row
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("📋 *Sizning ma'lumotlaringiz* (sahifa %d/%d)\n\n", page+1, totalPages))

	for i, entry := range entries {
		decrypted, err := cm.Decrypt(entry.EncryptedData, userKey)
		idx := startIdx + i + 1

		if err != nil {
			sb.WriteString(fmt.Sprintf("%d. *%s*: ❌ _xato_\n", idx, entry.Service))
			continue
		}

		// Format entry with copyable code block
		sb.WriteString(fmt.Sprintf("%d. *%s*\n", idx, entry.Service))

		// Split value into words for separate copy buttons
		words := strings.Fields(decrypted)
		if len(words) >= 2 {
			// Multiple words - show each on new line
			for j, word := range words {
				sb.WriteString(fmt.Sprintf("   └ `%s`", word))
				if j < len(words)-1 {
					sb.WriteString("\n")
				}
			}
		} else {
			// Single value
			sb.WriteString(fmt.Sprintf("   └ `%s`", decrypted))
		}
		sb.WriteString("\n\n")
	}

	sb.WriteString("_💡 Nusxa olish uchun `kod` ustiga bosing_\n")
	sb.WriteString("_⏰ 30 soniyadan so'ng yashiriladi_")

	// PAGINATION BUTTONS
	if totalPages > 1 {
		var navBtns []telebot.Btn

		if page > 0 {
			navBtns = append(navBtns, markup.Data("◀️ Oldingi", "list_page", strconv.Itoa(page-1)))
		}

		navBtns = append(navBtns, markup.Data(fmt.Sprintf("📄 %d/%d", page+1, totalPages), "list_refresh"))

		if page < totalPages-1 {
			navBtns = append(navBtns, markup.Data("Keyingi ▶️", "list_page", strconv.Itoa(page+1)))
		}

		rows = append(rows, markup.Row(navBtns...))
	}

	// Refresh button
	rows = append(rows, markup.Row(markup.Data("🔄 Yangilash", "list_refresh")))

	markup.Inline(rows...)
	return sb.String(), markup
}

// scheduleListExpiration hides list message after 30 seconds.
func scheduleListExpiration(b *telebot.Bot, msg *telebot.Message) {
	go func(m *telebot.Message) {
		time.Sleep(30 * time.Second)
		expiredText := "⏰ *Muddati tugadi*\n\n_Xavfsizlik sababli yashirildi._"
		if _, err := b.Edit(m, expiredText, telebot.ModeMarkdown); err != nil {
			log.Printf("Warning: Failed to edit expired list: %v", err)
		}
	}(msg)
}
