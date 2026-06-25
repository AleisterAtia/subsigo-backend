package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"github.com/sitepat/subsigo-backend/internal/models"
	"github.com/sitepat/subsigo-backend/pkg/database"
)

func main() {
	// Load .env (abaikan error jika dijalankan dengan env dari sistem, mis. Docker).
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  File .env tidak ditemukan, memakai environment variable sistem")
	}

	// Koneksi ke Neon PostgreSQL.
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("❌ Koneksi database gagal: %v", err)
	}

	// Jalankan migrasi skema (membuat tabel jika belum ada).
	if err := models.Migrate(db); err != nil {
		log.Fatalf("❌ Migrasi database gagal: %v", err)
	}
	log.Println("✅ Migrasi database selesai")

	app := fiber.New(fiber.Config{
		AppName: "SubsiGo Backend",
	})

	app.Use(recover.New())
	app.Use(logger.New())

	// Health check: sekaligus memverifikasi koneksi DB masih hidup.
	app.Get("/health", func(c *fiber.Ctx) error {
		sqlDB, err := database.DB.DB()
		if err != nil || sqlDB.Ping() != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "error",
				"db":     "down",
			})
		}
		return c.JSON(fiber.Map{
			"status": "ok",
			"db":     "up",
		})
	})

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server berjalan di port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("❌ Server gagal start: %v", err)
	}
}
