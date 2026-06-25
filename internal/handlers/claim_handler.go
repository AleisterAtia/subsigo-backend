package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sitepat/subsigo-backend/internal/middlewares"
	"github.com/sitepat/subsigo-backend/internal/services"
)

// ClaimHandler menangani endpoint klaim subsidi oleh petugas lapangan.
type ClaimHandler struct {
	claims *services.ClaimService
}

func NewClaimHandler(claims *services.ClaimService) *ClaimHandler {
	return &ClaimHandler{claims: claims}
}

type claimRequest struct {
	NFCUID    string `json:"nfc_uid"`
	Commodity string `json:"commodity"`
}

// Claim menangani POST /api/v1/claims.
//
// Penolakan bisnis (UID tak terdaftar, tidak layak, kuota habis) dikembalikan dengan
// HTTP 200 + body {"status":"rejected","message":...} agar aplikasi petugas cukup
// membaca field status. Hanya error teknis/validasi yang memakai kode 4xx/5xx.
func (h *ClaimHandler) Claim(c *fiber.Ctx) error {
	var req claimRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "format request tidak valid")
	}
	if req.NFCUID == "" || req.Commodity == "" {
		return fiber.NewError(fiber.StatusBadRequest, "nfc_uid dan commodity wajib diisi")
	}

	userID := c.Locals(middlewares.CtxUserID).(uuid.UUID)
	merchantName, _ := c.Locals(middlewares.CtxMerchant).(string)

	result, err := h.claims.Claim(userID, merchantName, req.NFCUID, req.Commodity)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCommodity) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, "gagal memproses klaim")
	}
	return c.JSON(result)
}
