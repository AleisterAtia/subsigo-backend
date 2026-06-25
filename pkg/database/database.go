package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB adalah instance koneksi global yang dipakai repository.
var DB *gorm.DB

// Connect membuka koneksi ke Neon PostgreSQL menggunakan DATABASE_URL dari environment,
// mengatur connection pool, lalu memverifikasi koneksi dengan Ping.
func Connect() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL belum di-set di environment/.env")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold: 500 * time.Millisecond,
				LogLevel:      logger.Warn,
				// "record not found" adalah alur normal (mis. UID tak terdaftar), jangan dianggap error.
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
		// Neon meng-host di luar negeri; biarkan timestamp dalam UTC agar konsisten.
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		return nil, fmt.Errorf("gagal membuka koneksi database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil handle sql.DB: %w", err)
	}

	// Tuning pool. Neon free-tier "scales to zero" saat idle, jadi koneksi
	// jangan disimpan terlalu lama agar tidak memakai koneksi yang sudah mati.
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(1 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("gagal ping ke database (cek connection string & password): %w", err)
	}

	DB = db
	log.Println("✅ Berhasil terhubung ke Neon PostgreSQL")
	return db, nil
}
