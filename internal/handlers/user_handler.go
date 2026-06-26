package handlers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sitepat/subsigo-backend/internal/models"
	"github.com/sitepat/subsigo-backend/internal/services"
)

// Batas panjang password. Maksimum 72 byte karena bcrypt memotong diam-diam di atas itu.
const (
	minPasswordLen = 6
	maxPasswordLen = 72
)

// UserHandler menangani manajemen akun (admin & petugas) — semua admin-only.
type UserHandler struct {
	users *services.UserService
}

func NewUserHandler(users *services.UserService) *UserHandler {
	return &UserHandler{users: users}
}

type createUserRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	Role         string `json:"role"`
	MerchantName string `json:"merchant_name"`
}

// CreateUser menangani POST /api/v1/admin/users.
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req createUserRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "format request tidak valid")
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" {
		return fiber.NewError(fiber.StatusBadRequest, "username wajib diisi")
	}
	if err := validatePassword(req.Password); err != nil {
		return err
	}
	if !models.IsValidRole(req.Role) {
		return fiber.NewError(fiber.StatusBadRequest, "role harus 'admin' atau 'merchant'")
	}

	user, err := h.users.CreateUser(req.Username, req.Password, req.Role, req.MerchantName)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fiber.NewError(fiber.StatusConflict, "username sudah dipakai")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "gagal membuat user")
	}
	return c.Status(fiber.StatusCreated).JSON(user)
}

// ListUsers menangani GET /api/v1/admin/users?search=&page=&limit=.
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	page, limit, offset := pageParams(c)
	users, total, err := h.users.ListUsers(c.Query("search"), limit, offset)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "gagal mengambil data user")
	}
	return paginated(c, users, page, limit, total)
}

type updateUserRequest struct {
	Role         *string `json:"role"`
	MerchantName *string `json:"merchant_name"`
	Password     *string `json:"password"`
	IsActive     *bool   `json:"is_active"`
}

// UpdateUser menangani PATCH /api/v1/admin/users/:id (semua field opsional).
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "id user tidak valid")
	}
	var req updateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "format request tidak valid")
	}
	if req.Role != nil && !models.IsValidRole(*req.Role) {
		return fiber.NewError(fiber.StatusBadRequest, "role harus 'admin' atau 'merchant'")
	}
	if req.Password != nil {
		if err := validatePassword(*req.Password); err != nil {
			return err
		}
	}

	user, err := h.users.UpdateUser(id, req.Role, req.MerchantName, req.Password, req.IsActive)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		case errors.Is(err, services.ErrNoUpdateFields):
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		default:
			return fiber.NewError(fiber.StatusInternalServerError, "gagal memperbarui user")
		}
	}
	return c.JSON(user)
}

// validatePassword memvalidasi panjang password.
func validatePassword(pw string) error {
	if len(pw) < minPasswordLen {
		return fiber.NewError(fiber.StatusBadRequest, "password minimal 6 karakter")
	}
	if len(pw) > maxPasswordLen {
		return fiber.NewError(fiber.StatusBadRequest, "password maksimal 72 karakter")
	}
	return nil
}
