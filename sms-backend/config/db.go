package config

import (
	"log"
	"os"
	"strings"
	"time"

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

	// Supabase (and most managed Postgres providers) require SSL.
	// Append sslmode=require if the caller hasn't specified one.
	if !strings.Contains(dsn, "sslmode") {
		if strings.Contains(dsn, "?") {
			dsn += "&sslmode=require"
		} else {
			dsn += "?sslmode=require"
		}
	}

	// Log which host we're actually connecting to (never log credentials).
	if idx := strings.Index(dsn, "@"); idx != -1 {
		hostPart := dsn[idx+1:]
		if end := strings.IndexAny(hostPart, "/?"); end != -1 {
			hostPart = hostPart[:end]
		}
		log.Printf("[db] Connecting to host: %s", hostPart)
	}

	for retries := 0; retries < 5; retries++ {
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger:      logger.Default.LogMode(logger.Warn),
			PrepareStmt: false, // required for PgBouncer / Supabase pooler
		})

		if err == nil {
			DB = db
			log.Println("Database connection established")
			return
		}

		log.Printf("Failed to connect to database (attempt %d/5): %v", retries+1, err)
		if retries < 4 {
			time.Sleep(3 * time.Second)
		}
	}

	log.Fatal("Failed to connect to database after 5 attempts")
}
