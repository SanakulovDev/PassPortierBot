package handlers

import (
	"os"

	"gopkg.in/telebot.v3"
)

// HandleAdd sends the "Add Password" WebApp button.
func HandleAdd() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		webAppURL := os.Getenv("WEBAPP_URL")
		if webAppURL == "" {
			webAppURL = "https://bot.sanakulov.uz/add_password.html"
		}

		menu := &telebot.ReplyMarkup{ResizeKeyboard: true}
		btnWebApp := menu.WebApp("➕ Parol Qo'shish", &telebot.WebApp{
			URL: webAppURL,
		})

		menu.Reply(menu.Row(btnWebApp))

		return c.Send("📝 Yangi parol qo'shish uchun pastdagi tugmani bosing:", menu)
	}
}
