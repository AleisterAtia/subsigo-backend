package handlers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sitepat/subsigo-backend/internal/models"
	"github.com/sitepat/subsigo-backend/internal/services"
)

// AdminHandler menangani endpoint admin (registrasi warga, kelayakan, kuota, monitoring).
type AdminHandler struct {
	admin *services.AdminService
}

func NewAdminHandler(admin *services.AdminService) *AdminHandler {
	return &AdminHandler{admin: admin}
}

type registerCitizenRequest struct {
	NIK    string `json:"nik"`
	NFCUID string `json:"nfc_uid"`
	Name   string `json:"name"`
}

// RegisterCitizen menangani POST /api/v1/admin/citizens.
func (h *AdminHandler) RegisterCitizen(c *fiber.Ctx) error {
	var req registerCitizenRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "format request tidak valid")
	}
	if !isValidNIK(req.NIK) {
		return fiber.NewError(fiber.StatusBadRequest, "NIK harus 16 digit angka")
	}
	if req.NFCUID == "" || req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "nfc_uid dan name wajib diisi")
	}

	citizen, err := h.admin.RegisterCitizen(req.NIK, req.NFCUID, req.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fiber.NewError(fiber.StatusConflict, "NIK atau NFC UID sudah terdaftar")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "gagal mendaftarkan warga")
	}
	return c.Status(fiber.StatusCreated).JSON(citizen)
}

type setEligibilityRequest struct {
	IsEligible *bool `json:"is_eligible"`
}

// SetEligibility menangani PATCH /api/v1/admin/citizens/:id/eligibility.
func (h *AdminHandler) SetEligibility(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "id warga tidak valid")
	}
	var req setEligibilityRequest
	if err := c.BodyParser(&req); err != nil || req.IsEligible == nil {
		return fiber.NewError(fiber.StatusBadRequest, "field is_eligible (true/false) wajib diisi")
	}

	if err := h.admin.SetEligibility(id, *req.IsEligible); err != nil {
		if errors.Is(err, services.ErrCitizenNotFound) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, "gagal memperbarui kelayakan")
	}
	return c.JSON(fiber.Map{"id": id, "is_eligible": *req.IsEligible})
}

type setQuotaRequest struct {
	Commodity  string `json:"commodity"`
	Period     string `json:"period"` // opsional, default bulan berjalan "YYYY-MM"
	QuotaTotal int    `json:"quota_total"`
}

// SetQuota menangani POST /api/v1/admin/citizens/:id/quotas.
func (h *AdminHandler) SetQuota(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "id warga tidak valid")
	}
	var req setQuotaRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "format request tidak valid")
	}
	if !models.IsValidCommodity(req.Commodity) {
		return fiber.NewError(fiber.StatusBadRequest, "jenis komoditas tidak valid")
	}
	if req.QuotaTotal < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "quota_total tidak boleh negatif")
	}
	if req.Period == "" {
		req.Period = time.Now().UTC().Format("2006-01")
	}

	quota, err := h.admin.SetQuota(id, req.Commodity, req.Period, req.QuotaTotal)
	if err != nil {
		if errors.Is(err, services.ErrCitizenNotFound) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, "gagal menetapkan kuota")
	}
	return c.Status(fiber.StatusCreated).JSON(quota)
}

// ListTransactions menangani GET /api/v1/admin/transactions?limit=N.
func (h *AdminHandler) ListTransactions(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 50)
	txs, err := h.admin.ListTransactions(limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "gagal mengambil transaksi")
	}
	return c.JSON(fiber.Map{"count": len(txs), "data": txs})
}

// isValidNIK memastikan NIK terdiri dari 16 digit angka.
func isValidNIK(nik string) bool {
	if len(nik) != 16 {
		return false
	}
	for _, ch := range nik {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}
