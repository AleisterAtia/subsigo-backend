# Deployment — Vercel (Go Framework Preset / Web Server)

Vercel mendeteksi project ini sebagai **Go web server** (bukan serverless function
per-file). "Go Framework Preset" mencari `go.mod` di root + entrypoint
`cmd/api/main.go`, lalu menjalankannya sebagai **server long-running**. Vercel
menyuntikkan env `PORT` dan server WAJIB listen di port itu.

> Karena itu kita TIDAK memakai `api/index.go` + `vercel.json` (model serverless
> function). Routing sepenuhnya ditangani Fiber; Vercel mem-proxy semua request
> ke server.

## Struktur yang relevan

```
cmd/api/main.go       # entrypoint yang dijalankan Vercel (web server)
internal/server/      # pembangunan app Fiber (route, middleware, DI)
internal/config/      # resolvePort(): pakai PORT (Vercel) -> APP_PORT -> 8080
cmd/migrate/          # migrasi skema (dijalankan terpisah, lihat di bawah)
cmd/seed/             # seeding user awal
```

## Setting project di Vercel

1. **Framework Preset**: pastikan = **Go** (Settings → General). Biasanya
   terdeteksi otomatis karena ada `cmd/api/main.go`.
2. **Root Directory**: root repo (tempat `go.mod` berada).
3. **Environment Variables** (Settings → Environment Variables):
   - `DATABASE_URL` → connection string **Pooled** dari Neon (host `...-pooler.neon.tech`).
   - `JWT_SECRET` → string acak yang kuat.
   - `JWT_EXPIRE_HOURS` (opsional, default 24).
   - `CORS_ALLOW_ORIGINS` → domain dashboard admin, mis. `https://<admin>.vercel.app`.
   - `DB_MAX_OPEN_CONNS` / `DB_MAX_IDLE_CONNS` (opsional).
   - JANGAN set `PORT` / `APP_PORT` di Vercel — `PORT` disuntik otomatis oleh Vercel.

## Migrasi & seeding (sekali, dari lokal/CI)

Vercel tidak menjalankan ini otomatis. Arahkan ke Neon yang sama:

```bash
go run ./cmd/migrate
go run ./cmd/seed
```

## Verifikasi setelah deploy

```
GET  https://<project>.vercel.app/health   -> {"status":"ok","db":"up"}
POST https://<project>.vercel.app/api/v1/auth/login
```

Catatan: path `/` tidak punya route (akan balas 404 JSON) — itu normal. Tes lewat `/health`.

## Catatan

- **Versi Go**: `go.mod` memakai `go 1.26.4`. Pastikan didukung build image Vercel;
  bila gagal, turunkan ke versi Go yang didukung.
- Server membaca `PORT` lewat `internal/config.resolvePort()`. Untuk dev lokal
  (`go run ./cmd/api`) tidak ada `PORT`, jadi memakai `APP_PORT`/8080.
