package config

import (
	"log"
	"os"
	"strconv"
)

// Config menampung konfigurasi aplikasi yang dibaca dari environment.
type Config struct {
	Port           string
	JWTSecret      string
	JWTExpireHours int
	// CORSAllowOrigins: daftar origin yang diizinkan memanggil API (dipisah koma),
	// terutama untuk dashboard admin Next.js. Default "*" untuk kemudahan dev.
	CORSAllowOrigins string
}

// Load membaca konfigurasi dari environment dan menghentikan aplikasi
// bila ada nilai wajib yang belum di-set.
func Load() *Config {
	cfg := &Config{
		Port:             getEnv("APP_PORT", "8080"),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		JWTExpireHours:   getEnvInt("JWT_EXPIRE_HOURS", 24),
		CORSAllowOrigins: getEnv("CORS_ALLOW_ORIGINS", "*"),
	}

	if cfg.JWTSecret == "" || cfg.JWTSecret == "ganti-dengan-secret-acak-yang-kuat" {
		log.Fatal("JWT_SECRET belum di-set dengan nilai yang aman di .env")
	}

	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
