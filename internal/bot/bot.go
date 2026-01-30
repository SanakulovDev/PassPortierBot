package bot

import (
	"log"
	"os"
	"time"

	"passportier-bot/internal/handlers"

	"gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
	"gorm.io/gorm"
)

// New creates and configures a new Telegram bot instance.
func New(db *gorm.DB) (*telebot.Bot, error) {
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
		return nil, err
	}

	b.Use(middleware.Logger())
	RegisterHandlers(b, db)
	SetCommands(b)

	return b, nil
}

// RegisterHandlers registers all bot command and message handlers.
func RegisterHandlers(b *telebot.Bot, db *gorm.DB) {
	b.Handle("/start", handlers.HandleStart())
	b.Handle("/unlock", handlers.HandleUnlock(b))
	b.Handle("/lock", handlers.HandleLock(b))
	b.Handle("/get", handlers.HandleGet(b, db))
	b.Handle("/list", handlers.HandleList(b, db))
	b.Handle(telebot.OnText, handlers.HandleText(b, db))

	// Register inline button callbacks
	handlers.RegisterListCallbacks(b, db)
}

// SetCommands registers bot commands with Telegram for the menu.
func SetCommands(b *telebot.Bot) {
	commands := []telebot.Command{
		{Text: "start", Description: "🚀 Botni ishga tushirish"},
		{Text: "unlock", Description: "🔓 Sessiyani ochish (30 daqiqa)"},
		{Text: "lock", Description: "🔒 Sessiyani yopish"},
		{Text: "list", Description: "📋 Barcha ma'lumotlarni ko'rish"},
		{Text: "get", Description: "🔍 Ma'lumot olish (masalan: /get instagram)"},
	}

	if err := b.SetCommands(commands); err != nil {
		log.Printf("Warning: Failed to set commands: %v", err)
	} else {
		log.Println("[BOT] Commands registered successfully")
	}
}
