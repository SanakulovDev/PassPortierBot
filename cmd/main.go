package main

import (
	"crypto/rand"
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
	"gorm.io/gorm/clause"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Initialize Database
	db := storage.InitDB()
	// Migrate new User table
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("User Migration Failed:", err)
	}

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
		// Check private chat only
		if c.Chat().Type != telebot.ChatPrivate {
			return nil // Silent ignore in groups
		}

		// Immediately delete user message for security
		if err := b.Delete(c.Message()); err != nil {
			log.Println("Warning: Failed to delete unlock message:", err)
		}

		// Manual parsing
		args := strings.SplitN(c.Text(), " ", 2)
		if len(args) < 2 {
			return c.Send("⚠️ Iltimos, maxfiy so'z kiriting! Misol: `/unlock mySecretPass`", telebot.ModeMarkdown)
		}
		passphrase := strings.TrimSpace(args[1])
		if passphrase == "" {
			return c.Send("⚠️ Iltimos, maxfiy so'z kiriting!", telebot.ModeMarkdown)
		}

		// 1. Get or Create User to fetch Salt
		var user models.User
		// Try to find user
		if err := db.First(&user, "telegram_id = ?", c.Sender().ID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// New User: Generate Salt
				salt := make([]byte, 16)
				if _, err := rand.Read(salt); err != nil {
					log.Printf("Salt gen error: %v", err)
					return c.Send("❌ Tizim xatosi (Salt generation).")
				}
				user = models.User{
					TelegramID: c.Sender().ID,
					Salt:       salt,
				}
				if err := db.Create(&user).Error; err != nil {
					log.Printf("User creation error: %v", err)
					return c.Send("❌ Foydalanuvchi yaratishda xatolik.")
				}
			} else {
				log.Printf("DB error fetching user: %v", err)
				return c.Send("❌ Tizim xatosi.")
			}
		}

		// 2. Derive Key using Argon2id
		// SECURITY: We use the user's unique salt and the passphrase
		key := crypto.DeriveKey(passphrase, user.Salt)

		// 3. Store in Vault
		vault.SetKey(c.Sender().ID, key)

		log.Printf("[DEBUG] User %d unlocked session. Key set (Argon2id).", c.Sender().ID)
		return c.Send("🔓 Sessiya ochildi! Sizning kalitingiz Argon2id bilan shifrlanib, 30 daqiqa RAMda saqlanadi.")
	})

	// /get [service] - Retrieve and decrypt password
	b.Handle("/get", func(c telebot.Context) error {
		// Immediately delete user message for security
		if err := b.Delete(c.Message()); err != nil {
			log.Println("Warning: Failed to delete get message:", err)
		}

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

		// Normalize input: lowercase service names
		text = strings.ToLower(text)

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
					return retrievePassword(c, b, db, serviceName)
				}
				return savePassword(c, b, db, serviceName, data)
			}

			// Case 2: GET (#key)
			return retrievePassword(c, b, db, serviceName)
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

	// Save to DB using UPSERT (PostgreSQL specific)
	entry := models.PasswordEntry{
		UserID:        c.Sender().ID,
		Service:       serviceName,
		EncryptedData: encrypted,
	}

	// upsert: conflict on (user_id, service) -> update encrypted_data
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "service"}},
		DoUpdates: clause.AssignmentColumns([]string{"encrypted_data", "updated_at"}),
	}).Create(&entry).Error; err != nil {
		// Log error and cleanup loading message
		if delErr := b.Delete(msg); delErr != nil {
			log.Println("Warning: Failed to delete saving message:", delErr)
		}
		log.Printf("DB Error: %v", err)
		return c.Send("❌ Bazaga saqlash xatosi.")
	}

	if err := b.Delete(msg); err != nil {
		log.Println("Warning: Failed to delete saving message:", err)
	}
	return c.Send(fmt.Sprintf("✅ *%s* saqlandi!", serviceName), telebot.ModeMarkdown)
}

func retrievePassword(c telebot.Context, b *telebot.Bot, db *gorm.DB, serviceName string) error {
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

	// Send password and schedule deletion after 5 minutes
	msgText := fmt.Sprintf("🔑 *%s*\n\n`%s`\n\n⚠️ _Bu xabar xavfsizlik uchun 5 daqiqadan so'ng o'chiriladi._", entry.Service, string(decrypted))
	sentMsg, err := b.Send(c.Sender(), msgText, telebot.ModeMarkdown)
	if err != nil {
		return err
	}

	// Launch goroutine to auto-delete the response
	go func(m *telebot.Message) {
		time.Sleep(5 * time.Minute)
		if err := b.Delete(m); err != nil {
			log.Printf("Warning: Failed to auto-delete response message (id %d): %v", m.ID, err)
		}
	}(sentMsg)

	return nil
}
