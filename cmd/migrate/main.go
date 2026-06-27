// Command migrate menjalankan migrasi skema database secara terpisah dari server API,
// lalu menjalankan backfill idempoten untuk memetakan data subsidi lama ke model layanan
// generik yang baru. Jalankan dengan: go run ./cmd/migrate (aman dijalankan berulang).
package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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

	// 1. AutoMigrate: membuat tabel/kolom/index baru (services, service_eligibilities,
	//    service_quotas, dan kolom transactions.service_id/service_code/metadata).
	//    Tidak pernah menghapus kolom/tabel lama — itu ditangani backfill di bawah.
	if err := models.Migrate(db); err != nil {
		log.Fatalf("❌ Migrasi gagal: %v", err)
	}
	log.Println("✅ AutoMigrate selesai")

	// 2. Backfill: idempotent (aman diulang) & sadar-skema (melewati langkah yang tak
	//    relevan pada database baru yang belum punya data/kolom lama).
	if err := backfill(db); err != nil {
		log.Fatalf("❌ Backfill gagal: %v", err)
	}
	log.Println("✅ Backfill selesai")

	// Verifikasi: tampilkan daftar tabel yang ada di database.
	var tables []string
	db.Raw(`SELECT table_name FROM information_schema.tables
	        WHERE table_schema = current_schema() AND table_type = 'BASE TABLE'
	        ORDER BY table_name`).Scan(&tables)
	log.Printf("📋 Tabel di database: %v", tables)
}

// backfill memindahkan data dari skema subsidi lama ke model layanan generik.
// Semua langkah idempotent (ON CONFLICT DO NOTHING / WHERE ... IS NULL) sehingga
// menjalankan migrate dua kali tidak menggandakan atau menggagalkan apa pun.
func backfill(db *gorm.DB) error {
	// (a) Seed layanan bawaan dari SATU sumber kebenaran (models.SeedServices).
	//     Layanan harus ada LEBIH DULU karena langkah b–d memetakan data ke service_id.
	for i := range models.SeedServices() {
		s := models.SeedServices()[i]
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "code"}},
			DoNothing: true,
		}).Create(&s).Error; err != nil {
			return fmt.Errorf("seed layanan %q: %w", s.Code, err)
		}
	}

	mig := db.Migrator()

	// (b) Kelayakan: dari kolom lama citizens.is_eligible × layanan subsidi.
	//     Hanya pada database hasil migrasi dari skema lama (kolom is_eligible masih ada).
	if mig.HasColumn(&models.Citizen{}, "is_eligible") {
		err := db.Exec(`
			INSERT INTO service_eligibilities (citizen_id, service_id, is_eligible, created_at, updated_at)
			SELECT c.id, s.id, c.is_eligible, now(), now()
			FROM citizens c
			JOIN services s ON s.code IN (?, ?)
			ON CONFLICT (citizen_id, service_id) DO NOTHING
		`, models.ServiceCodeLPG3KG, models.ServiceCodePertalite).Error
		if err != nil {
			return fmt.Errorf("backfill kelayakan: %w", err)
		}
		log.Println("   ↳ kelayakan subsidi di-backfill dari citizens.is_eligible")
	}

	// (c) Kuota: salin subsidy_quotas -> service_quotas, memetakan commodity -> service_id.
	if mig.HasTable("subsidy_quotas") {
		err := db.Exec(`
			INSERT INTO service_quotas (citizen_id, service_id, service_code, period, quota_total, quota_remaining, created_at, updated_at)
			SELECT sq.citizen_id, s.id, s.code, sq.period, sq.quota_total, sq.quota_remaining, sq.created_at, sq.updated_at
			FROM subsidy_quotas sq
			JOIN services s ON s.code = sq.commodity
			ON CONFLICT (citizen_id, service_id, period) DO NOTHING
		`).Error
		if err != nil {
			return fmt.Errorf("backfill kuota: %w", err)
		}
		log.Println("   ↳ subsidy_quotas disalin ke service_quotas")
	}

	// (d) Transaksi: isi service_id/service_code dari kolom lama commodity, lalu lepas
	//     NOT NULL pada commodity. KRITIS: kolom commodity lama NOT NULL tanpa default —
	//     tanpa langkah ini, setiap INSERT klaim baru (yang tak lagi mengisi commodity)
	//     akan melanggar constraint. AutoMigrate tidak akan melakukan ALTER ini.
	if mig.HasColumn(&models.Transaction{}, "commodity") {
		err := db.Exec(`
			UPDATE transactions t
			SET service_id = s.id, service_code = s.code
			FROM services s
			WHERE s.code = t.commodity AND t.service_id IS NULL
		`).Error
		if err != nil {
			return fmt.Errorf("backfill transaksi: %w", err)
		}
		if err := db.Exec(`ALTER TABLE transactions ALTER COLUMN commodity DROP NOT NULL`).Error; err != nil {
			return fmt.Errorf("lepas NOT NULL transactions.commodity: %w", err)
		}
		log.Println("   ↳ transactions.service_id di-backfill & commodity dijadikan nullable")
	}

	return nil
}
