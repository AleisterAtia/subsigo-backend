# Deployment — Vercel Serverless

Backend ini berjalan sebagai **satu Vercel Serverless Function** (`api/index.go`).
Semua request di-rewrite ke fungsi itu (`vercel.json`), lalu ditangani oleh aplikasi
Fiber dari paket `internal/server`. Untuk dev lokal masih bisa `go run ./cmd/api`.

## Struktur yang relevan

```
api/index.go          # entry-point serverless (func Handler) -> adaptor.FiberApp
vercel.json           # rewrite semua path /(.*) -> /api/index
internal/server/      # pembangunan app Fiber (dipakai lokal & serverless)
cmd/api/              # server HTTP untuk dev lokal
cmd/migrate/          # migrasi skema (dijalankan terpisah, lihat di bawah)
cmd/seed/             # seeding user awal
```

## Langkah deploy

1. **Push repo ke GitHub**, lalu di Vercel: *New Project* → import repo ini.
   Vercel otomatis mendeteksi runtime Go dari `go.mod` + folder `api/`.

2. **Set Environment Variables** di Vercel (Project Settings → Environment Variables):
   - `DATABASE_URL` → gunakan connection string **Pooled** dari Neon
     (host `...-pooler.neon.tech`). Ini penting agar koneksi tidak cepat habis
     saat banyak instance serverless aktif.
   - `JWT_SECRET` → secret acak yang kuat.
   - `JWT_EXPIRE_HOURS` (opsional, default 24).
   - `CORS_ALLOW_ORIGINS` → domain dashboard admin, mis. `https://<admin>.vercel.app`.
   - `DB_MAX_OPEN_CONNS` / `DB_MAX_IDLE_CONNS` (opsional).
   - `APP_PORT` TIDAK perlu di Vercel.

3. **Jalankan migrasi & seeding** (sekali, dari mesin lokal/CI yang menunjuk Neon
   yang sama — Vercel tidak menjalankan ini otomatis):
   ```bash
   go run ./cmd/migrate
   go run ./cmd/seed
   ```

4. **Deploy** (otomatis setiap push ke branch, sesuai PRD). Verifikasi:
   ```
   GET  https://<project>.vercel.app/health            -> {"status":"ok","db":"up"}
   POST https://<project>.vercel.app/api/v1/auth/login
   ```

## Catatan / hal yang perlu dicek

- **Versi Go**: `go.mod` memakai `go 1.26.4`. Pastikan versi ini didukung oleh build
  image Vercel saat deploy; bila gagal, turunkan ke versi Go yang didukung Vercel.
- **Cold start**: koneksi DB diinisialisasi lazily sekali per instance dan dipakai
  ulang pada warm start. Jika init gagal (mis. Neon baru bangun dari scale-to-zero),
  request berikutnya otomatis mencoba lagi.
- **Path routing**: Fiber tetap melihat path asli (mis. `/api/v1/claims`) karena
  rewrite Vercel mempertahankan URL asli ke fungsi. Health check ada di `/health`.
