package main

import (
	"context"
	"log"

	"passportier-bot/internal/api"
	"passportier-bot/internal/bot"
	"passportier-bot/internal/models"
	"passportier-bot/internal/security"
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

	// Initialize Redis and SessionManager
	redisClient, err := security.NewRedisClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	sessionManager := security.NewSessionManager(redisClient)

	// Initialize and start bot
	b, err := bot.New(db, sessionManager)
	if err != nil {
		log.Fatal(err)
	}

	// Remove webhook before polling
	if err := b.RemoveWebhook(); err != nil {
		log.Printf("Warning: Failed to remove webhook: %v", err)
	}

	log.Println("PassPortierBot is running...")
	
	// Start API server for Web App
	apiServer := api.NewServer(db, sessionManager)
	go func() {
		if err := apiServer.Start(":8080"); err != nil {
			log.Printf("API server error: %v", err)
		}
	}()
	
	b.Start()
}
