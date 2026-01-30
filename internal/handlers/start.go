package handlers

import "gopkg.in/telebot.v3"

// HandleStart returns the /start command handler with welcome message.
func HandleStart() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		msg := "👋 <b>Assalomu alaykum, PassPortierBot-ga xush kelibsiz!</b>\n\n" +
			"Men sizning parollaringizni xavfsiz saqlashga yordam beraman. 🔐\n" +
			"Mening ishlash prinsipim <b>Zero-Knowledge</b> texnologiyasiga asoslangan: sizning maxfiy kalitingiz faqat vaqtinchalik xotirada (RAM) saqlanadi.\n\n" +
			"🚀 <b>Ishni boshlash uchun:</b>\n" +
			"1. <code>/unlock [maxfiy_so'z]</code> - Sessiyani ochish (kalit 30 daqiqa RAMda turadi).\n" +
			"2. Login/parol yuboring (yozma, rasm yoki ovozli).\n" +
			"3. <code>/get [xizmat_nomi]</code> - Parolni olish.\n\n" +
			"🔒 <b>Xavfsizlik:</b> Har bir ma'lumot AES-256-GCM shifrlash usuli bilan himoyalangan."
		return c.Send(msg, telebot.ModeHTML)
	}
}
