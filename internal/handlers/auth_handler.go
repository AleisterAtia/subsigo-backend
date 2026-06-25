package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/sitepat/subsigo-backend/internal/services"
)

// AuthHandler menangani endpoint autentikasi.
type AuthHandler struct {
	auth *services.AuthService
}

func NewAuthHandler(auth *services.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login memproses POST /api/v1/auth/login.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "format request tidak valid")
	}
	if req.Username == "" || req.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "username dan password wajib diisi")
	}

	tok, user, err := h.auth.Login(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, "terjadi kesalahan saat login")
	}

	return c.JSON(fiber.Map{
		"token": tok,
		"user": fiber.Map{
			"id":            user.ID,
			"username":      user.Username,
			"role":          user.Role,
			"merchant_name": user.MerchantName,
		},
	})
}
