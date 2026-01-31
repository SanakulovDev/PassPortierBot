package user

import (
	"fmt"
	"strconv"

	"passportier-bot/internal/models"

	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
)

// HandleSettings renders the settings menu.
func HandleSettings() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		menu := &telebot.ReplyMarkup{}
		
		btnImmed := menu.Data("🔒 Immediately", "autolock", "0")
		btn5Min := menu.Data("⏱ 5 Mins", "autolock", "300")
		btn1Hour := menu.Data("⏳ 1 Hour", "autolock", "3600")

		menu.Inline(
			menu.Row(btnImmed),
			menu.Row(btn5Min, btn1Hour),
		)

		return c.Send("⚙️ *Settings*\n\nChoose Auto-Lock Duration:", menu)
	}
}

// HandleAutoLockCallback updates the user's session TTL preference.
func HandleAutoLockCallback(db *gorm.DB) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		ttlStr := c.Data()
		ttl, err := strconv.ParseInt(ttlStr, 10, 64)
		if err != nil {
			return c.Respond(&telebot.CallbackResponse{Text: "Invalid option"})
		}

		userID := c.Sender().ID
		
		// Update user setting in DB
		if err := db.Model(&models.User{}).Where("telegram_id = ?", userID).Update("session_ttl", ttl).Error; err != nil {
			return c.Respond(&telebot.CallbackResponse{Text: "Failed to update settings"})
		}
		
		msg := fmt.Sprintf("✅ Auto-Lock set to: %s", formatDuration(ttl))
		c.Edit(msg)
		return c.Respond(&telebot.CallbackResponse{Text: "Settings saved"})
	}
}

func formatDuration(seconds int64) string {
	if seconds == 0 {
		return "Immediately"
	} else if seconds < 3600 {
		return fmt.Sprintf("%d mins", seconds/60)
	}
	return fmt.Sprintf("%d hour(s)", seconds/3600)
}
