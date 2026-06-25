package main

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
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

func main() {
	// Load .env (abaikan error jika dijalankan dengan env dari sistem, mis. Docker).
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  File .env tidak ditemukan, memakai environment variable sistem")
	}

	cfg := config.Load()

	// Koneksi ke Neon PostgreSQL.
	// Catatan: migrasi skema TIDAK dijalankan di sini (lihat cmd/migrate),
	// agar aman saat di-deploy dengan banyak replika di Docker.
	if _, err := database.Connect(); err != nil {
		log.Fatalf("❌ Koneksi database gagal: %v", err)
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

	log.Printf("🚀 Server berjalan di port %s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("❌ Server gagal start: %v", err)
	}
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
