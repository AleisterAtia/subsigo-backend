// Command seed mengisi data awal (user admin & petugas) untuk testing.
// Jalankan dengan: go run ./cmd/seed  (idempotent, aman dijalankan berulang)
package main

import (
	"errors"
	"log"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"github.com/sitepat/subsigo-backend/internal/models"
	"github.com/sitepat/subsigo-backend/pkg/database"
	"github.com/sitepat/subsigo-backend/pkg/hash"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  File .env tidak ditemukan, memakai environment variable sistem")
	}

	if _, err := database.Connect(); err != nil {
		log.Fatalf("❌ Koneksi database gagal: %v", err)
	}
	db := database.DB

	seedUser(db, "admin", "admin123", models.RoleAdmin, "")
	seedUser(db, "petugas1", "petugas123", models.RoleMerchant, "SPBU 34-401 Merdeka")

	log.Println("✅ Seeding selesai")
}

func seedUser(db *gorm.DB, username, password, role, merchant string) {
	var existing models.User
	err := db.Where("username = ?", username).First(&existing).Error
	if err == nil {
		log.Printf("↷ User '%s' sudah ada, dilewati", username)
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Fatalf("❌ Gagal mengecek user '%s': %v", username, err)
	}

	hashed, err := hash.Hash(password)
	if err != nil {
		log.Fatalf("❌ Gagal hash password: %v", err)
	}

	u := models.User{
		Username:     username,
		PasswordHash: hashed,
		Role:         role,
		MerchantName: merchant,
	}
	if err := db.Create(&u).Error; err != nil {
		log.Fatalf("❌ Gagal membuat user '%s': %v", username, err)
	}
	log.Printf("✅ User '%s' (role: %s) dibuat", username, role)
}
