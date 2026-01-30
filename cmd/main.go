package main

import (
	"log"

	"passportier-bot/internal/bot"
	"passportier-bot/internal/models"
	"passportier-bot/internal/storage"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Initialize Database
	db := storage.InitDB()

	// Run migrations
	if err := db.AutoMigrate(&models.User{}, &models.PasswordEntry{}); err != nil {
		log.Printf("Migration Failed: %v", err)
	}

	// Initialize and start bot
	b, err := bot.New(db)
	if err != nil {
		log.Fatal(err)
	}

	// Remove webhook before polling
	if err := b.RemoveWebhook(); err != nil {
		log.Printf("Warning: Failed to remove webhook: %v", err)
	}

	log.Println("PassPortierBot is running...")
	b.Start()
}
