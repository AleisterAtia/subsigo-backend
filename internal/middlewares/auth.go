package middlewares

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/sitepat/subsigo-backend/pkg/token"
)

// Kunci penyimpanan data user di context Fiber (c.Locals).
const (
	CtxUserID = "userID"
	CtxRole   = "role"
)

// RequireAuth memvalidasi JWT dari header Authorization: Bearer <token>.
// Bila valid, menyimpan userID & role ke c.Locals untuk handler berikutnya.
func RequireAuth(tm *token.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return fiber.NewError(fiber.StatusUnauthorized, "token tidak ditemukan")
		}

		tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		claims, err := tm.Parse(tokenStr)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}

		c.Locals(CtxUserID, claims.UserID)
		c.Locals(CtxRole, claims.Role)
		return c.Next()
	}
}

// RequireRole membatasi akses hanya untuk role tertentu (mis. admin).
// Harus dipasang SETELAH RequireAuth.
func RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Locals(CtxRole) != role {
			return fiber.NewError(fiber.StatusForbidden, "akses ditolak untuk role ini")
		}
		return c.Next()
	}
}
