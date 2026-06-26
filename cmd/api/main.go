// Command api menjalankan server HTTP untuk pengembangan lokal (go run ./cmd/api).
//
// Catatan: di PRODUKSI backend dijalankan sebagai Vercel Serverless Function
// (lihat api/index.go), BUKAN lewat command ini. Keduanya memakai aplikasi Fiber
// yang sama dari paket internal/server, jadi perilakunya konsisten.
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
