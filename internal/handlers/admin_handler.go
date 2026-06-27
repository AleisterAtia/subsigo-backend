package handlers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sitepat/subsigo-backend/internal/models"
	"github.com/sitepat/subsigo-backend/internal/repositories"
	"github.com/sitepat/subsigo-backend/internal/services"
)

// AdminHandler menangani endpoint admin (registrasi warga, kelayakan, kuota, monitoring).
type AdminHandler struct {
	admin *services.AdminService
}

func NewAdminHandler(admin *services.AdminService) *AdminHandler {
	return &AdminHandler{admin: admin}
}

// ListCitizens menangani GET /api/v1/admin/citizens?search=&page=&limit=.
func (h *AdminHandler) ListCitizens(c *fiber.Ctx) error {
	page, limit, offset := pageParams(c)
	citizens, total, err := h.admin.ListCitizens(c.Query("search"), limit, offset)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "gagal mengambil data warga")
	}
	return paginated(c, citizens, page, limit, total)
}

// GetCitizen menangani GET /api/v1/admin/citizens/:id (detail + kuota).
func (h *AdminHandler) GetCitizen(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "id warga tidak valid")
	}
	citizen, err := h.admin.GetCitizen(id)
	if err != nil {
		if errors.Is(err, services.ErrCitizenNotFound) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, "gagal mengambil data warga")
	}
	return c.JSON(citizen)
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
	// ServiceCode opsional: bila kosong, kelayakan diterapkan ke SEMUA layanan aktif
	// yang membutuhkan kelayakan (kompat dengan tombol kelayakan global lama).
	ServiceCode string `json:"service_code"`
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

	if err := h.admin.SetEligibility(id, req.ServiceCode, *req.IsEligible); err != nil {
		switch {
		case errors.Is(err, services.ErrCitizenNotFound):
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		case errors.Is(err, services.ErrServiceNotFound):
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		default:
			return fiber.NewError(fiber.StatusInternalServerError, "gagal memperbarui kelayakan")
		}
	}
	return c.JSON(fiber.Map{"id": id, "service_code": req.ServiceCode, "is_eligible": *req.IsEligible})
}

type setQuotaRequest struct {
	ServiceCode string `json:"service_code"`
	Commodity   string `json:"commodity"` // alias lama; dipakai bila service_code kosong (kompat)
	Period      string `json:"period"`    // opsional, default bulan berjalan "YYYY-MM"
	QuotaTotal  int    `json:"quota_total"`
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
	serviceCode := req.ServiceCode
	if serviceCode == "" {
		serviceCode = req.Commodity // kompat: terima field "commodity" lama
	}
	if serviceCode == "" {
		return fiber.NewError(fiber.StatusBadRequest, "service_code wajib diisi")
	}
	if req.QuotaTotal < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "quota_total tidak boleh negatif")
	}
	if req.Period == "" {
		// Default ke periode berjalan menurut WIB, konsisten dengan derivasi periode
		// saat klaim (lihat ClaimService) — agar kuota yang baru di-set langsung terpakai.
		req.Period = models.CurrentPeriod()
	} else if !models.IsValidPeriod(req.Period) {
		return fiber.NewError(fiber.StatusBadRequest, "period harus berformat YYYY-MM (mis. 2026-06)")
	}

	quota, err := h.admin.SetQuota(id, serviceCode, req.Period, req.QuotaTotal)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCitizenNotFound):
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		case errors.Is(err, services.ErrServiceNotFound), errors.Is(err, services.ErrServiceNotQuota):
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		default:
			return fiber.NewError(fiber.StatusInternalServerError, "gagal menetapkan kuota")
		}
	}
	return c.Status(fiber.StatusCreated).JSON(quota)
}

// ListTransactions menangani GET /api/v1/admin/transactions dengan filter & pagination.
// Query: page, limit, status, commodity, user_id, merchant_name, from, to (YYYY-MM-DD, WIB).
func (h *AdminHandler) ListTransactions(c *fiber.Ctx) error {
	page, limit, offset := pageParams(c)

	f := repositories.TransactionFilter{
		MerchantName: c.Query("merchant_name"),
		Limit:        limit,
		Offset:       offset,
	}

	if status := c.Query("status"); status != "" {
		if !models.IsValidTxStatus(status) {
			return fiber.NewError(fiber.StatusBadRequest, "status harus 'success' atau 'rejected'")
		}
		f.Status = status
	}
	// Terima ?service_code= (baru) atau ?commodity= (lama, kompat) sebagai filter layanan.
	// Tidak divalidasi terhadap daftar tetap karena layanan kini dinamis — kode tak dikenal
	// hanya menghasilkan 0 baris, bukan error.
	if svc := c.Query("service_code"); svc != "" {
		f.ServiceCode = svc
	} else if commodity := c.Query("commodity"); commodity != "" {
		f.ServiceCode = commodity
	}
	if uid := c.Query("user_id"); uid != "" {
		parsed, err := uuid.Parse(uid)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "user_id tidak valid")
		}
		f.UserID = &parsed
	}
	if from := c.Query("from"); from != "" {
		t, err := time.ParseInLocation("2006-01-02", from, models.WIB)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "from harus berformat YYYY-MM-DD")
		}
		f.From = &t
	}
	if to := c.Query("to"); to != "" {
		t, err := time.ParseInLocation("2006-01-02", to, models.WIB)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "to harus berformat YYYY-MM-DD")
		}
		// Eksklusif di tengah malam WIB berikutnya agar seluruh hari 'to' ikut tercakup.
		end := t.AddDate(0, 0, 1)
		f.To = &end
	}

	txs, total, err := h.admin.ListTransactions(f)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "gagal mengambil transaksi")
	}
	return paginated(c, txs, page, limit, total)
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
