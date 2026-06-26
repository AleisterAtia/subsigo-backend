package config

import (
	"errors"
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

// Load membaca konfigurasi dari environment dan mengembalikan error
// bila ada nilai wajib yang belum di-set. Sengaja TIDAK memakai log.Fatal:
// di serverless (Vercel) os.Exit akan membuat fungsi crash
// (FUNCTION_INVOCATION_FAILED) tanpa pesan yang berguna.
func Load() (*Config, error) {
	cfg := &Config{
		Port:             resolvePort(),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		JWTExpireHours:   getEnvInt("JWT_EXPIRE_HOURS", 24),
		CORSAllowOrigins: getEnv("CORS_ALLOW_ORIGINS", "*"),
	}

	if cfg.JWTSecret == "" || cfg.JWTSecret == "ganti-dengan-secret-acak-yang-kuat" {
		return nil, errors.New("JWT_SECRET belum di-set dengan nilai yang aman (cek Environment Variables di Vercel)")
	}

	return cfg, nil
}

// resolvePort menentukan port HTTP yang dipakai server.
// Vercel (Go Framework Preset) MENYUNTIKKAN env PORT dengan port acak yang WAJIB
// dipakai — kalau server listen di port lain, Vercel menganggap gagal start.
// Untuk dev lokal: pakai APP_PORT, fallback 8080.
func resolvePort() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return getEnv("APP_PORT", "8080")
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
