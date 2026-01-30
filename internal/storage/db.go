package storage

import (
	"fmt"
	"os"
	"passportier-bot/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Bazaga ulanib bo'lmadi!")
	}

	// Jadvalni avtomatik yaratish
	if err := db.AutoMigrate(&models.PasswordEntry{}); err != nil {
		panic("Migratsiya xatosi: " + err.Error())
	}
	return db
}
