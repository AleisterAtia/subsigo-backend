// Command migrate menjalankan migrasi skema database secara terpisah dari server API.
// Jalankan dengan: go run ./cmd/migrate
package main

import (
	"log"

	"github.com/joho/godotenv"

	"github.com/sitepat/subsigo-backend/internal/models"
	"github.com/sitepat/subsigo-backend/pkg/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  File .env tidak ditemukan, memakai environment variable sistem")
	}

	db, err := database.Connect()
	if err != nil {
		log.Fatalf("❌ Koneksi database gagal: %v", err)
	}

	if err := models.Migrate(db); err != nil {
		log.Fatalf("❌ Migrasi gagal: %v", err)
	}
	log.Println("✅ Migrasi selesai")

	// Verifikasi: tampilkan daftar tabel yang ada di database.
	var tables []string
	db.Raw(`SELECT table_name FROM information_schema.tables
	        WHERE table_schema = current_schema() AND table_type = 'BASE TABLE'
	        ORDER BY table_name`).Scan(&tables)
	log.Printf("📋 Tabel di database: %v", tables)
}
