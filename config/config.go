package config

import (
	"fmt"
	"log"
	"os"

	"github.com/21johnh21/route-planner-api/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB() *gorm.DB {
	// Load .env if present
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"), os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	log.Println("Initializing database connection...")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Running database migrations...")
	if err := db.AutoMigrate(&models.User{}, &models.Trail{}); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	return db
}
