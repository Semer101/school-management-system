package config

import (
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"

	"sms-backend/models"
)

// EnsureDefaultAdmin creates a default admin user if no users exist in the database.
// This is a bootstrap mechanism so the first login on a fresh deployment works.
// Run it after ConnectDB() and AutoMigrate.
func EnsureDefaultAdmin() {
	if DB == nil {
		log.Println("[seed] DB not initialized, skipping default admin")
		return
	}

	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		log.Printf("[seed] %d users already exist, skipping default admin", count)
		return
	}

	email := os.Getenv("DEFAULT_ADMIN_EMAIL")
	if email == "" {
		email = "admin@school.et"
	}
	password := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	if password == "" {
		password = "Admin@1234"
	}
	name := os.Getenv("DEFAULT_ADMIN_NAME")
	if name == "" {
		name = "System Administrator"
	}
	phone := os.Getenv("DEFAULT_ADMIN_PHONE")
	if phone == "" {
		phone = "0911234567"
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Printf("[seed] Failed to hash default admin password: %v", err)
		return
	}

	admin := models.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
		Role:     models.RoleAdmin,
		Phone:    phone,
		IsActive: true,
	}

	if err := DB.Create(&admin).Error; err != nil {
		log.Printf("[seed] Failed to create default admin: %v", err)
		return
	}

	log.Printf("[seed] Created default admin user: %s / %s", email, password)
	log.Printf("[seed] CHANGE THIS PASSWORD AFTER FIRST LOGIN")
}