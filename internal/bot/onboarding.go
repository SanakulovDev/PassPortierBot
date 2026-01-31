package bot

import (
	"gopkg.in/telebot.v3"
)

// HandleOnboarding sends a rich media welcome message.
func HandleOnboarding() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		// Rich media: Animation (GIF) or Video
		// Using a placeholder URL. In production, use file_id like "CgACAgIAAxkBA..."
		// For now, sending a text with WebApp button explanation.
		
		caption := `🔒 *PassPortier Bot ga xush kelibsiz!*

Men sizning ma'lumotlaringizni *Zero-Knowledge* tamoyili asosida himoyalayman.

🛡 *Bu qanday ishlaydi?*
1. Siz /unlock [so'z] orqali sessiya ochasiz.
2. Bu so'z *faqat RAM da* (xotirada) saqlanadi.
3. Ma'lumotlaringiz shifrlanib bazaga yoziladi.
4. Sessiya tugagach, kalit butunlay o'chiriladi.

*Men hatto sizning parolingizni bilmayman!*

Boshlash uchun pastdagi tugmani bosing 👇`

		menu := &telebot.ReplyMarkup{}
		btnWebApp := menu.WebApp("➕ Parol Qo'shish", &telebot.WebApp{
			URL: "c", // Replace with actual URL
		})
		btnSettings := menu.Text("⚙️ Sozlamalar")

		menu.Reply(
			menu.Row(btnWebApp),
			menu.Row(btnSettings),
		)
		
		// Ideally send Animation.
		// return c.Send(&telebot.Animation{File: telebot.FromURL("https://media.giphy.com/media/l0HlS0SlpQxY4B1wA/giphy.gif"), Caption: caption}, menu)
		
		// Sending Photo for reliability in this demo context
		return c.Send(&telebot.Photo{
			File:    telebot.FromURL("https://cdn-icons-png.flaticon.com/512/3064/3064197.png"),
			Caption: caption,
		}, 
		// menu options
		menu,
		telebot.ModeMarkdown,
		)
	}
}
