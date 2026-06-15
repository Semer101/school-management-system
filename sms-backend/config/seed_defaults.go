package config

import (
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"

	"sms-backend/models"
)

// defaultUser defines credentials for a bootstrap user account.
type defaultUser struct {
	Name     string
	Email    string
	Password string
	Phone    string
	Role     string
}

// EnsureDefaultUsers creates default users for all 4 roles if the users table is empty.
// This is a bootstrap mechanism so the first login on a fresh deployment works.
// Run it after ConnectDB() and AutoMigrate.
func EnsureDefaultUsers() {
	if DB == nil {
		log.Println("[seed] DB not initialized, skipping default users")
		return
	}

	log.Println("[seed] Checking for default users...")

	// Always ensure default users exist with correct passwords
	defaults := []defaultUser{
		{
			Name:     "System Administrator",
			Email:    envOrDefault("DEFAULT_ADMIN_EMAIL", "admin@school.et"),
			Password: envOrDefault("DEFAULT_ADMIN_PASSWORD", "Admin@1234"),
			Phone:    "0911234567",
			Role:     models.RoleAdmin,
		},
		{
			Name:     "Default Teacher",
			Email:    "teacher1@school.et",
			Password: "Teacher@1234",
			Phone:    "0911100001",
			Role:     models.RoleTeacher,
		},
		{
			Name:     "Default Student",
			Email:    "student1@school.et",
			Password: "Student@1234",
			Phone:    "0911500001",
			Role:     models.RoleStudent,
		},
		{
			Name:     "Default Parent",
			Email:    "parent1@school.et",
			Password: "Parent@1234",
			Phone:    "0944100001",
			Role:     models.RoleParent,
		},
	}

	for _, u := range defaults {
		hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
		var existing models.User
		result := DB.Where("email = ?", u.Email).First(&existing)
		if result.Error != nil {
			user := models.User{
				Name:     u.Name,
				Email:    u.Email,
				Password: string(hashedPwd),
				Role:     u.Role,
				Phone:    u.Phone,
				IsActive: true,
			}
			if err := DB.Create(&user).Error; err != nil {
				log.Printf("[seed] Failed to create %s: %v", u.Email, err)
			} else {
				log.Printf("[seed] Created %s: %s", u.Role, u.Email)
			}
		} else {
			// Always update password to ensure it matches expected value
			if err := DB.Model(&existing).Updates(map[string]any{
				"password":   string(hashedPwd),
				"is_active":  true,
				"name":       u.Name,
				"phone":      u.Phone,
			}).Error; err != nil {
				log.Printf("[seed] Failed to update %s: %v", u.Email, err)
			} else {
				log.Printf("[seed] Updated %s password", u.Email)
			}
		}
	}

	log.Println("[seed] Default users ready")
}

// envOrDefault returns the env var value or a fallback default.
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
