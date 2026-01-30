package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"passportier-bot/internal/crypto"
	"passportier-bot/internal/models"
	"passportier-bot/internal/storage"
	"passportier-bot/internal/vault"

	"github.com/joho/godotenv"
	"gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
	"gorm.io/gorm"
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
		OnError: func(err error, c telebot.Context) {
			if c != nil {
				log.Printf("[ERROR] Update %v failed: %v", c.Update().ID, err)
			} else {
				log.Printf("[ERROR] Bot error: %v", err)
			}
		},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Middleware
	b.Use(middleware.Logger())

	// /start - Welcome message
	b.Handle("/start", func(c telebot.Context) error {
		msg := "👋 <b>Assalomu alaykum, PassPortierBot-ga xush kelibsiz!</b>\n\n" +
			"Men sizning parollaringizni xavfsiz saqlashga yordam beraman. 🔐\n" +
			"Mening ishlash prinsipim <b>Zero-Knowledge</b> texnologiyasiga asoslangan: sizning maxfiy kalitingiz faqat vaqtinchalik xotirada (RAM) saqlanadi.\n\n" +
			"🚀 <b>Ishni boshlash uchun:</b>\n" +
			"1. <code>/unlock [maxfiy_so'z]</code> - Sessiyani ochish (kalit 30 daqiqa RAMda turadi).\n" +
			"2. Login/parol yuboring (yozma, rasm yoki ovozli).\n" +
			"3. <code>/get [xizmat_nomi]</code> - Parolni olish.\n\n" +
			"🔒 <b>Xavfsizlik:</b> Har bir ma'lumot AES-256-GCM shifrlash usuli bilan himoyalangan."
		return c.Send(msg, telebot.ModeHTML)
	})

	// /unlock [passphrase] - Generate 32-byte AES Key from passphrase using SHA-256 and store in RAM
	b.Handle("/unlock", func(c telebot.Context) error {
		// Manual parsing to be safe
		args := strings.SplitN(c.Text(), " ", 2)
		if len(args) < 2 {
			return c.Send("⚠️ Iltimos, maxfiy so'z kiriting! Misol: `/unlock mySecretPass`", telebot.ModeMarkdown)
		}
		passphrase := strings.TrimSpace(args[1])
		if passphrase == "" {
			return c.Send("⚠️ Iltimos, maxfiy so'z kiriting!", telebot.ModeMarkdown)
		}

		// Foydalanuvchi kiritgan so'zdan 32 baytlik kalit yasash (SHA-256)
		hash := sha256.Sum256([]byte(passphrase))
		vault.SetKey(c.Sender().ID, hash[:])

		log.Printf("[DEBUG] User %d unlocked session. Passphrase len: %d. Key set.", c.Sender().ID, len(passphrase))
		return c.Send("🔓 Sessiya ochildi! Siz kiritgan so'zdan maxsus 32-baytlik shifrlash kaliti yasald.\nEndi 30 daqiqa davomida ma'lumot yuborishingiz mumkin.")
	})

	// /get [service] - Retrieve and decrypt password
	b.Handle("/get", func(c telebot.Context) error {
		log.Printf("[DEBUG] /get request from User %d", c.Sender().ID)
		userKey, ok := vault.GetKey(c.Sender().ID)
		if !ok {
			log.Printf("[DEBUG] User %d session NOT found", c.Sender().ID)
			return c.Send("🔒 Sessiya yopiq. Avval `/unlock [so'z]` buyrug'ini yuboring.", telebot.ModeMarkdown)
		}

		serviceName := strings.TrimSpace(c.Message().Payload)
		if serviceName == "" {
			// Try manual parse if Payload fails
			args := strings.SplitN(c.Text(), " ", 2)
			if len(args) >= 2 {
				serviceName = strings.TrimSpace(args[1])
			}
		}

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
			log.Printf("[ERROR] Decryption failed for User %d Service %s: %v", c.Sender().ID, serviceName, err)
			return c.Send("❌ Shifrni ochib bo'lmadi. Kalit noto'g'ri bo'lishi mumkin.")
		}

		return c.Send(fmt.Sprintf("🔑 *%s*\n\n%s", entry.Service, string(decrypted)), telebot.ModeMarkdown)
	})

	// Handle Text
	b.Handle(telebot.OnText, func(c telebot.Context) error {
		// Xavfsizlik uchun foydalanuvchi xabarini o'chiramiz (defer orqali funksiya oxirida)
		defer func() {
			if err := b.Delete(c.Message()); err != nil {
				log.Println("Warning: Failed to delete user message:", err)
			}
		}()

		text := strings.TrimSpace(c.Text())

		// Check if it starts with #
		if strings.HasPrefix(text, "#") {
			// Remove #
			cleanText := strings.TrimPrefix(text, "#")

			// Split by whitespace (newline, space, etc.)
			// We use regex to find the first whitespace separator
			re := regexp.MustCompile(`\s+`)
			parts := re.Split(strings.TrimSpace(cleanText), 2)

			serviceName := strings.TrimSpace(parts[0])
			if serviceName == "" {
				return c.Send("⚠️ Iltimos, xizmat nomini hash bilan yozing. Misol: `#instagram` yoki `#instagram parol123`")
			}

			// Case 1: SAVE (#key data)
			if len(parts) > 1 {
				data := strings.TrimSpace(parts[1])
				if data == "" {
					// Edge case where there is just whitespace
					return retrievePassword(c, db, serviceName)
				}
				return savePassword(c, b, db, serviceName, data)
			}

			// Case 2: GET (#key)
			return retrievePassword(c, db, serviceName)
		}

		// If not starting with #, maybe just chat or ignore
		return c.Send("⚠️ Ma'lumot saqlash uchun `#xizmat_nomi ma'lumot` ko'rinishida yozing.\nOlish uchun esa shunchaki `#xizmat_nomi` deb yozing.")
	})

	log.Println("PassPortierBot is running...")

	// Explicitly remove webhook to ensure polling works
	if err := b.RemoveWebhook(); err != nil {
		log.Printf("Warning: Failed to remove webhook: %v", err)
	}

	b.Start()
}

func savePassword(c telebot.Context, b *telebot.Bot, db *gorm.DB, serviceName, data string) error {
	userKey, ok := vault.GetKey(c.Sender().ID)
	if !ok {
		return c.Send("🔒 Sessiya yopiq. Avval `/unlock [so'z]` buyrug'ini yuboring.")
	}

	msg, err := b.Send(c.Sender(), "⏳ Saqlanmoqda...")
	if err != nil {
		log.Println("Send error:", err)
	}

	// Encrypt the data as is (it's up to user format now)
	encrypted, err := crypto.Encrypt([]byte(data), userKey)
	if err != nil {
		if err := b.Delete(msg); err != nil {
			log.Println("Warning: Failed to delete saving message:", err)
		}
		return c.Send("❌ Shifrlash xatosi.")
	}

	// Save to DB
	entry := models.PasswordEntry{
		UserID:        c.Sender().ID,
		Service:       serviceName,
		EncryptedData: encrypted,
	}

	if err := db.Create(&entry).Error; err != nil {
		if err := b.Delete(msg); err != nil {
			log.Println("Warning: Failed to delete saving message:", err)
		}
		return c.Send("❌ Bazaga saqlash xatosi.")
	}

	if err := b.Delete(msg); err != nil {
		log.Println("Warning: Failed to delete saving message:", err)
	}
	return c.Send(fmt.Sprintf("✅ *%s* saqlandi!", serviceName), telebot.ModeMarkdown)
}

func retrievePassword(c telebot.Context, db *gorm.DB, serviceName string) error {
	userKey, ok := vault.GetKey(c.Sender().ID)
	if !ok {
		return c.Send("🔒 Sessiya yopiq. Avval `/unlock [so'z]` buyrug'ini yuboring.")
	}

	var entry models.PasswordEntry
	// Case-insensitive search
	result := db.Where("user_id = ? AND LOWER(service) LIKE ?", c.Sender().ID, "%"+strings.ToLower(serviceName)+"%").First(&entry)
	if result.Error != nil {
		return c.Send(fmt.Sprintf("❌ *%s* bo'yicha ma'lumot topilmadi.", serviceName), telebot.ModeMarkdown)
	}

	decrypted, err := crypto.Decrypt(entry.EncryptedData, userKey)
	if err != nil {
		log.Printf("[ERROR] Decryption failed for User %d Service %s: %v", c.Sender().ID, serviceName, err)
		return c.Send("❌ Shifrni ochib bo'lmadi. Kalit noto'g'ri bo'lishi mumkin.")
	}

	return c.Send(fmt.Sprintf("🔑 *%s*\n\n`%s`", entry.Service, string(decrypted)), telebot.ModeMarkdown)
}
