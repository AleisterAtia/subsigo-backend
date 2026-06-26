// Command api adalah satu-satunya entry-point web server, dipakai untuk dev lokal
// (go run ./cmd/api) DAN di produksi Vercel. Vercel "Go Framework Preset" mendeteksi
// cmd/api/main.go, menjalankannya sebagai web server long-running, lalu menyuntikkan
// env PORT yang WAJIB di-bind (lihat internal/config.resolvePort & DEPLOYMENT.md).
package main

import (
	"log"

	"github.com/sitepat/subsigo-backend/internal/server"
)

func main() {
	app, cfg, err := server.New()
	if err != nil {
		log.Fatalf("❌ Gagal inisialisasi aplikasi: %v", err)
	}

	log.Printf("🚀 Server berjalan di port %s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("❌ Server gagal start: %v", err)
	}
}
