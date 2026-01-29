package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"passportier-bot/internal/ai"
	"passportier-bot/internal/crypto"
	"passportier-bot/internal/models"
	"passportier-bot/internal/storage"
	"passportier-bot/internal/vault"

	"github.com/joho/godotenv"
	"gopkg.in/telebot.v3"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Initialize Database
	db := storage.InitDB()

	// Initialize Bot
	pref := telebot.Settings{
		Token:  os.Getenv("BOT_TOKEN"),
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Middleware to check if user has an active session (for protected commands if needed)
	// However, distinct handlers will check it manually for finer control.

	// /unlock [key] - Set Master Key in RAM
	b.Handle("/unlock", func(c telebot.Context) error {
		key := c.Message().Payload
		if len(key) != 32 {
			return c.Send("⚠️ Master Key roppa-rosa 32 ta belgidan iborat bo'lishi kerak!")
		}
		vault.SetKey(c.Sender().ID, []byte(key))
		return c.Send("🔓 Sessiya ochildi! Endi 30 daqiqa davomida ma'lumot yuborishingiz mumkin.")
	})

	// /get [service] - Retrieve and decrypt password
	b.Handle("/get", func(c telebot.Context) error {
		userKey, ok := vault.GetKey(c.Sender().ID)
		if !ok {
			return c.Send("🔒 Sessiya yopiq. Avval /unlock [key] buyrug'ini yuboring.")
		}

		serviceName := strings.TrimSpace(c.Message().Payload)
		if serviceName == "" {
			return c.Send("⚠️ Qaysi xizmatni qidiryapsiz? Misol: /get google")
		}

		var entry models.PasswordEntry
		// Case-insensitive search
		result := db.Where("user_id = ? AND LOWER(service) LIKE ?", c.Sender().ID, "%"+strings.ToLower(serviceName)+"%").First(&entry)
		if result.Error != nil {
			return c.Send("❌ Topilmadi.")
		}

		decrypted, err := crypto.Decrypt(entry.EncryptedData, userKey)
		if err != nil {
			return c.Send("❌ Shifrni ochib bo'lmadi. Kalit noto'g'ri bo'lishi mumkin.")
		}

		return c.Send(fmt.Sprintf("🔑 *%s*\n\n%s", entry.Service, string(decrypted)), telebot.ModeMarkdown)
	})

	// Helper function to process content via AI
	processContent := func(c telebot.Context, mimeType string, data []byte) error {
		userKey, ok := vault.GetKey(c.Sender().ID)
		if !ok {
			return c.Send("🔒 Sessiya yopiq. Avval /unlock [key] buyrug'ini yuboring.")
		}

		msg, err := b.Send(c.Sender(), "⏳ AI tahlil qilmoqda...")
		if err != nil {
			log.Println("Send error:", err)
		}

		cred, err := ai.ParseInput(context.Background(), mimeType, data)
		if err != nil {
			b.Delete(msg)
			return c.Send(fmt.Sprintf("❌ AI xatosi: %v", err))
		}

		// Encrypt the full JSON credential as one blob
		credJSON := fmt.Sprintf(`{"service":"%s", "login":"%s", "password":"%s", "note":"%s"}`,
			cred.Service, cred.Login, cred.Password, cred.Note)

		encrypted, err := crypto.Encrypt([]byte(credJSON), userKey)
		if err != nil {
			b.Delete(msg)
			return c.Send("❌ Shifrlash xatosi.")
		}

		// Save to DB
		entry := models.PasswordEntry{
			UserID:        c.Sender().ID,
			Service:       cred.Service,
			EncryptedData: encrypted,
		}

		if err := db.Create(&entry).Error; err != nil {
			b.Delete(msg)
			return c.Send("❌ Bazaga saqlash xatosi.")
		}

		b.Delete(msg)
		return c.Send(fmt.Sprintf("✅ *%s* saqlandi!\nLogin: `%s`", cred.Service, cred.Login), telebot.ModeMarkdown)
	}

	// Handle Text
	b.Handle(telebot.OnText, func(c telebot.Context) error {
		// Ignore commands
		if strings.HasPrefix(c.Text(), "/") {
			return nil
		}
		return processContent(c, "text/plain", []byte(c.Text()))
	})

	// Handle Photo
	b.Handle(telebot.OnPhoto, func(c telebot.Context) error {
		file := c.Message().Photo.File
		reader, err := b.File(&file)
		if err != nil {
			return c.Send("❌ Rasmni yuklab bo'lmadi.")
		}
		defer reader.Close()

		data, err := io.ReadAll(reader)
		if err != nil {
			return c.Send("❌ Rasmni o'qib bo'lmadi.")
		}

		return processContent(c, "image/jpeg", data)
	})

	// Handle Voice
	b.Handle(telebot.OnVoice, func(c telebot.Context) error {
		file := c.Message().Voice.File
		reader, err := b.File(&file)
		if err != nil {
			return c.Send("❌ Ovozli xabarni yuklab bo'lmadi.")
		}
		defer reader.Close()

		data, err := io.ReadAll(reader)
		if err != nil {
			return c.Send("❌ Ovozli xabarni o'qib bo'lmadi.")
		}

		// AI usually supports audio/mp3 or audio/wav. Telegram voice is usually OGG/Opus.
		// Gemini supports various audio formats. We'll verify mimetype or convert if needed.
		// For now, passing generic audio type or "audio/ogg".
		return processContent(c, "audio/ogg", data)
	})

	log.Println("PassPortierBot is running...")
	b.Start()
}
