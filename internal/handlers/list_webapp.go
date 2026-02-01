package handlers

import (
	"os"

	"gopkg.in/telebot.v3"
)

// HandleListWebApp sends a button to open the password manager Web App.
func HandleListWebApp() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		webAppURL := os.Getenv("WEBAPP_LIST_URL")
		if webAppURL == "" {
			webAppURL = "https://bot.sanakulov.uz/passwords.html"
		}

		menu := &telebot.ReplyMarkup{ResizeKeyboard: true}
		btnWebApp := menu.WebApp("📋 Parollarim", &telebot.WebApp{
			URL: webAppURL,
		})

		menu.Reply(menu.Row(btnWebApp))

		return c.Send("🔐 *Parol menejerni ochish uchun pastdagi tugmani bosing:*\n\n_Eslatma: Avval /unlock qilganingizga ishonch hosil qiling!_", menu, telebot.ModeMarkdown)
	}
}
