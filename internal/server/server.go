// Package server membangun aplikasi Fiber lengkap (DB, dependency injection, route).
// Sengaja dipisah dari main agar dipakai BERSAMA oleh dua entry-point:
//   - server lokal untuk dev: cmd/api (go run ./cmd/api)
//   - fungsi serverless Vercel: api/index.go
//
// Dengan begitu konfigurasi route & middleware tidak terduplikasi di dua tempat.
package server

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/sitepat/subsigo-backend/internal/config"
	"github.com/sitepat/subsigo-backend/internal/handlers"
	"github.com/sitepat/subsigo-backend/internal/middlewares"
	"github.com/sitepat/subsigo-backend/internal/models"
	"github.com/sitepat/subsigo-backend/internal/repositories"
	"github.com/sitepat/subsigo-backend/internal/services"
	"github.com/sitepat/subsigo-backend/pkg/database"
	"github.com/sitepat/subsigo-backend/pkg/token"
)

// New memuat konfigurasi, membuka koneksi database, merangkai seluruh dependency,
// lalu mengembalikan aplikasi Fiber yang siap melayani request beserta Config-nya.
func New() (*fiber.App, *config.Config, error) {
	// Muat .env bila ada. Dilewati tanpa error di Vercel: di sana environment
	// variable disuntik oleh platform, bukan dari file .env.
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}

	// Koneksi ke Neon PostgreSQL. Migrasi skema TIDAK dijalankan di sini
	// (lihat cmd/migrate) agar aman saat banyak instance serverless hidup bersamaan.
	if _, err := database.Connect(); err != nil {
		return nil, nil, err
	}
	db := database.DB

	// --- Dependency injection ---
	tm := token.NewManager(cfg.JWTSecret, cfg.JWTExpireHours)

	// Repositories
	userRepo := repositories.NewUserRepository(db)
	citizenRepo := repositories.NewCitizenRepository(db)
	quotaRepo := repositories.NewQuotaRepository(db)
	txRepo := repositories.NewTransactionRepository(db)

	// Services
	authSvc := services.NewAuthService(userRepo, tm)
	adminSvc := services.NewAdminService(citizenRepo, quotaRepo, txRepo)
	claimSvc := services.NewClaimService(db)

	// Handlers
	authHandler := handlers.NewAuthHandler(authSvc)
	adminHandler := handlers.NewAdminHandler(adminSvc)
	claimHandler := handlers.NewClaimHandler(claimSvc)

	app := fiber.New(fiber.Config{
		AppName:      "SubsiGo Backend",
		ErrorHandler: jsonErrorHandler,
	})
	app.Use(recover.New())
	app.Use(logger.New())
	// CORS diperlukan agar dashboard admin (Next.js, dijalankan di browser)
	// dapat memanggil API ini. Tidak memakai cookie/credentials — token dikirim
	// lewat header Authorization — sehingga origin "*" pun aman untuk dev.
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSAllowOrigins,
		AllowMethods: "GET,POST,PATCH,PUT,DELETE,OPTIONS",
		AllowHeaders: "Authorization,Content-Type",
	}))

	// Health check: sekaligus memverifikasi koneksi DB masih hidup.
	app.Get("/health", func(c *fiber.Ctx) error {
		sqlDB, err := database.DB.DB()
		if err != nil || sqlDB.Ping() != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "error", "db": "down"})
		}
		return c.JSON(fiber.Map{"status": "ok", "db": "up"})
	})

	// --- Rute API v1 ---
	api := app.Group("/api/v1")
	api.Post("/auth/login", authHandler.Login)

	// Rute yang membutuhkan JWT valid.
	protected := api.Group("", middlewares.RequireAuth(tm))
	protected.Get("/me", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user_id":       c.Locals(middlewares.CtxUserID).(uuid.UUID),
			"role":          c.Locals(middlewares.CtxRole).(string),
			"merchant_name": c.Locals(middlewares.CtxMerchant),
		})
	})

	// Rute admin (hanya role admin).
	admin := protected.Group("/admin", middlewares.RequireRole(models.RoleAdmin))
	admin.Post("/citizens", adminHandler.RegisterCitizen)
	admin.Patch("/citizens/:id/eligibility", adminHandler.SetEligibility)
	admin.Post("/citizens/:id/quotas", adminHandler.SetQuota)
	admin.Get("/transactions", adminHandler.ListTransactions)

	// Rute klaim subsidi (hanya role merchant/petugas lapangan).
	claims := protected.Group("/claims", middlewares.RequireRole(models.RoleMerchant))
	claims.Post("", claimHandler.Claim)

	return app, cfg, nil
}

// jsonErrorHandler memformat semua error (termasuk fiber.NewError) menjadi JSON.
func jsonErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	msg := "terjadi kesalahan internal"
	var fe *fiber.Error
	if errors.As(err, &fe) {
		code = fe.Code
		msg = fe.Message
	}
	return c.Status(code).JSON(fiber.Map{"error": msg})
}
