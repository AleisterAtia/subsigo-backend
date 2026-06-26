package database

import (
	"fmt"
	"log"
	"os"
	"strconv"
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
		// Terjemahkan error driver ke error GORM (mis. gorm.ErrDuplicatedKey) agar mudah ditangani.
		TranslateError: true,
	})
	if err != nil {
		return nil, fmt.Errorf("gagal membuka koneksi database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil handle sql.DB: %w", err)
	}

	// Tuning pool. Di lingkungan SERVERLESS (Vercel) banyak instance bisa hidup
	// bersamaan dan masing-masing memegang pool sendiri — total koneksi = jumlah
	// instance × MaxOpenConns. Karena itu pool dijaga KECIL dan WAJIB memakai
	// DATABASE_URL "pooled" (host berakhiran "-pooler") dari Neon (PgBouncer)
	// agar koneksi tidak cepat habis. Nilainya bisa di-override lewat env.
	// Neon juga "scales to zero" saat idle, jadi koneksi jangan disimpan terlalu lama.
	sqlDB.SetMaxOpenConns(envInt("DB_MAX_OPEN_CONNS", 5))
	sqlDB.SetMaxIdleConns(envInt("DB_MAX_IDLE_CONNS", 2))
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(1 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("gagal ping ke database (cek connection string & password): %w", err)
	}

	DB = db
	log.Println("✅ Berhasil terhubung ke Neon PostgreSQL")
	return db, nil
}

// envInt membaca integer dari environment; mengembalikan def bila kosong/invalid.
func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
