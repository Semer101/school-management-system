package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() {
	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// safer for production debugging
		Logger: logger.Default.LogMode(logger.Warn),

		// IMPORTANT: avoids prepared statement issues on Supabase/Render
		PrepareStmt: false,
	})

	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	log.Println("Database connection established")
	DB = db
}
