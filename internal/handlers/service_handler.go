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

// ServiceHandler menangani manajemen katalog layanan (admin-only).
type ServiceHandler struct {
	svc *services.ServiceService
}

func NewServiceHandler(svc *services.ServiceService) *ServiceHandler {
	return &ServiceHandler{svc: svc}
}

// ListServices menangani GET /api/v1/admin/services (semua layanan, aktif & nonaktif).
func (h *ServiceHandler) ListServices(c *fiber.Ctx) error {
	list, err := h.svc.List()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "gagal mengambil daftar layanan")
	}
	// Bungkus dalam { data: [...] } agar konsisten dengan endpoint list lain.
	return c.JSON(fiber.Map{"data": list})
}

type createServiceRequest struct {
	Code            string `json:"code"`
	Name            string `json:"name"`
	Kind            string `json:"kind"`
	DefaultEligible bool   `json:"default_eligible"`
}

// CreateService menangani POST /api/v1/admin/services.
func (h *ServiceHandler) CreateService(c *fiber.Ctx) error {
	var req createServiceRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "format request tidak valid")
	}
	code := normalizeServiceCode(req.Code)
	name := strings.TrimSpace(req.Name)
	if !isValidServiceCode(code) {
		return fiber.NewError(fiber.StatusBadRequest, "code wajib 2-32 karakter huruf besar/angka/underscore (mis. LPG_3KG)")
	}
	if name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "nama layanan wajib diisi")
	}
	if !models.IsValidServiceKind(req.Kind) {
		return fiber.NewError(fiber.StatusBadRequest, "kind harus 'quota', 'eligibility', atau 'log'")
	}

	svc, err := h.svc.Create(code, name, req.Kind, req.DefaultEligible)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fiber.NewError(fiber.StatusConflict, "code layanan sudah dipakai")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "gagal membuat layanan")
	}
	return c.Status(fiber.StatusCreated).JSON(svc)
}

type updateServiceRequest struct {
	Name            *string `json:"name"`
	Kind            *string `json:"kind"`
	DefaultEligible *bool   `json:"default_eligible"`
	IsActive        *bool   `json:"is_active"`
}

// UpdateService menangani PATCH /api/v1/admin/services/:id (semua field opsional).
// Code sengaja tidak bisa diubah.
func (h *ServiceHandler) UpdateService(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "id layanan tidak valid")
	}
	var req updateServiceRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "format request tidak valid")
	}
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			return fiber.NewError(fiber.StatusBadRequest, "nama layanan tidak boleh kosong")
		}
		req.Name = &trimmed
	}
	if req.Kind != nil && !models.IsValidServiceKind(*req.Kind) {
		return fiber.NewError(fiber.StatusBadRequest, "kind harus 'quota', 'eligibility', atau 'log'")
	}

	svc, err := h.svc.Update(id, req.Name, req.Kind, req.DefaultEligible, req.IsActive)
	if err != nil {
		if errors.Is(err, services.ErrServiceNotFound) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, "gagal memperbarui layanan")
	}
	return c.JSON(svc)
}

// normalizeServiceCode menyeragamkan code: trim + huruf besar.
func normalizeServiceCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// isValidServiceCode memastikan code 2-32 karakter [A-Z0-9_].
func isValidServiceCode(code string) bool {
	if len(code) < 2 || len(code) > 32 {
		return false
	}
	for _, ch := range code {
		if !((ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_') {
			return false
		}
	}
	return true
}
