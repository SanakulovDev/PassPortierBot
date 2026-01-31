package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"passportier-bot/internal/security"
	"passportier-bot/internal/services"

	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
)

// WebAppPayload represents the JSON structure sent by the Mini App.
type WebAppPayload struct {
	Service string `json:"service"`
	Data    string `json:"data"`
}

// HandleWebApp processes data sent from the Mini App.
func HandleWebApp(b *telebot.Bot, db *gorm.DB, sm *security.SessionManager) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		// WebApp data comes via Service message or text?
		// telebot.OnWebApp is for OnData from WebApp.
		
		webAppData := c.Message().WebAppData
		if webAppData == nil {
			return nil
		}

		var payload WebAppPayload
		if err := json.Unmarshal([]byte(webAppData.Data), &payload); err != nil {
			log.Printf("Failed to unmarshal webapp data: %v", err)
			return c.Send("❌ Ma'lumot formati noto'g'ri.")
		}
		
		if payload.Service == "" || payload.Data == "" {
			return c.Send("⚠️ Xizmat nomi va ma'lumot bo'sh bo'lmasligi kerak.")
		}

		// Save to DB
		// Verify session is active
		if _, err := sm.GetSession(context.Background(), c.Sender().ID); err != nil {
			return c.Send("🔒 Sessiya yopiq! Iltimos, avval `/unlock` qiling va qayta urinib ko'ring.")
		}

		if err := services.SavePassword(context.Background(), db, sm, c.Sender().ID, payload.Service, payload.Data); err != nil {
			log.Printf("Failed to save from WebApp: %v", err)
			return c.Send("❌ Saqlashda xatolik yuz berdi.")
		}

		return c.Send(fmt.Sprintf("✅ *%s* muvaffaqiyatli saqlandi!", payload.Service), telebot.ModeMarkdown)
	}
}
