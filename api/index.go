// Package handler adalah entry-point Vercel Serverless Function untuk SELURUH API.
//
// Runtime Go Vercel mendeteksi file di dalam folder /api dan memanggil fungsi
// bernama Handler dengan signature net/http. Semua path diarahkan ke fungsi
// tunggal ini lewat "rewrites" di vercel.json, lalu diteruskan ke aplikasi Fiber
// melalui adaptor net/http — sehingga seluruh routing tetap ditangani Fiber.
package handler

import (
	"net/http"
	"sync"

	"github.com/gofiber/fiber/v2/middleware/adaptor"

	"github.com/sitepat/subsigo-backend/internal/server"
)

var (
	mu          sync.Mutex
	httpHandler http.HandlerFunc
)

// Handler adalah titik masuk yang dipanggil runtime Vercel pada setiap request.
func Handler(w http.ResponseWriter, r *http.Request) {
	h, err := build()
	if err != nil {
		// Gagal init (mis. DATABASE_URL salah, atau Neon masih "cold").
		http.Error(w, `{"error":"service unavailable: `+err.Error()+`"}`, http.StatusServiceUnavailable)
		return
	}
	h(w, r)
}

// build menginisialisasi aplikasi sekali per instance (warm start memakai ulang
// hasilnya). Bila init gagal, handler TIDAK di-cache sehingga request berikutnya
// mencoba lagi — berguna saat database Neon sedang bangun dari "scale to zero".
func build() (http.HandlerFunc, error) {
	mu.Lock()
	defer mu.Unlock()
	if httpHandler != nil {
		return httpHandler, nil
	}
	app, _, err := server.New()
	if err != nil {
		return nil, err
	}
	httpHandler = adaptor.FiberApp(app)
	return httpHandler, nil
}
