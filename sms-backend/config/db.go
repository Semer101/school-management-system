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
	godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	log.Println("Database connection established")
	DB = db
}

// PreMigrate ensures schema is in a state that AutoMigrate can handle cleanly.
// if the refresh_tokens table already exists but is missing the constraint,
// we ADD it here so GORM can safely drop-then-recreate it during AutoMigrate.
// If the table doesn't exist yet, AutoMigrate creates everything fresh and the
// DROP never runs at all.
func PreMigrate(db *gorm.DB) error {
	return db.Exec(`
		DO $$
		BEGIN
			-- Only act if the table exists (first run creates it from scratch via AutoMigrate)
			IF EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = current_schema()
				  AND table_name   = 'refresh_tokens'
			) THEN
				-- Ensure the constraint GORM will try to DROP actually exists
				IF NOT EXISTS (
					SELECT 1 FROM information_schema.table_constraints
					WHERE table_schema    = current_schema()
					  AND table_name      = 'refresh_tokens'
					  AND constraint_name = 'uni_refresh_tokens_user_id'
				) THEN
					ALTER TABLE refresh_tokens
					ADD CONSTRAINT uni_refresh_tokens_user_id UNIQUE (user_id);
				END IF;
			END IF;
		END;
		$$;
	`).Error
}
