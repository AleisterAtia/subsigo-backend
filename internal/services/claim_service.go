package services

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/sitepat/subsigo-backend/internal/models"
)

// ErrInvalidCommodity dikembalikan saat kode layanan tidak dikenal ATAU nonaktif.
// Nama variabel dipertahankan agar penanganan error di handler (dan jalur error di
// aplikasi mobile) tidak berubah — kompatibilitas API.
var ErrInvalidCommodity = errors.New("jenis layanan tidak valid")

// ClaimResult adalah hasil pemrosesan klaim yang dikirim ke aplikasi petugas.
// Bentuk JSON-nya SENGAJA tidak berubah dari versi subsidi agar mobile tetap kompatibel.
type ClaimResult struct {
	Status         string    `json:"status"`          // "success" | "rejected"
	Message        string    `json:"message"`         // pesan/alasan untuk ditampilkan
	QuotaRemaining int       `json:"quota_remaining"` // sisa kuota setelah klaim (0 untuk layanan tanpa kuota)
	TransactionID  uuid.UUID `json:"transaction_id"`
}

// ClaimService menangani logika menjalankan satu layanan lewat tap e-KTP.
// Menyimpan *gorm.DB karena alur kritis (kuota) dijalankan dalam satu transaksi
// dengan row-locking.
type ClaimService struct {
	db *gorm.DB
}

func NewClaimService(db *gorm.DB) *ClaimService {
	return &ClaimService{db: db}
}

// Claim memproses satu penggunaan layanan secara atomik & aman terhadap race condition.
// serviceCode adalah Service.Code (aplikasi mobile mengirimnya di field "commodity").
//
// Alur bercabang sesuai Service.Kind:
//   - log         : cari warga (tolak bila tak terdaftar) -> catat sukses.
//   - eligibility : cari warga -> cek kelayakan -> catat.
//   - quota       : cari warga -> cek kelayakan -> KUNCI baris kuota periode berjalan
//                   -> pastikan sisa > 0 -> potong -> catat (alur subsidi lama, kini per service_id).
//
// Setiap penolakan tetap dicatat sebagai transaksi (sesuai PRD: catat gagal & sukses).
func (s *ClaimService) Claim(userID uuid.UUID, merchantName, nfcUID, serviceCode string) (*ClaimResult, error) {
	// Resolusi layanan DI LUAR transaksi (operasi baca). Tak dikenal/nonaktif -> 400,
	// TIDAK dicatat — paritas dengan perilaku lama "komoditas tidak valid".
	var service models.Service
	err := s.db.Where("code = ? AND is_active = ?", serviceCode, true).First(&service).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidCommodity
	}
	if err != nil {
		return nil, err
	}

	// Periode kuota berjalan ("YYYY-MM" menurut WIB) — hanya relevan untuk Kind=quota.
	period := models.CurrentPeriod()

	var result ClaimResult

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// (a) Cari warga berdasarkan UID e-KTP.
		var citizen models.Citizen
		err := tx.Where("nfc_uid = ?", nfcUID).First(&citizen).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return s.record(tx, &result, &service, models.TxStatusRejected, models.ReasonNotRegistered,
				nil, nfcUID, userID, merchantName, 0)
		}
		if err != nil {
			return err
		}

		// (b) Cek kelayakan untuk layanan yang membutuhkannya (quota & eligibility).
		if service.Kind == models.ServiceKindQuota || service.Kind == models.ServiceKindEligibility {
			eligible, err := s.isEligible(tx, citizen.ID, &service)
			if err != nil {
				return err
			}
			if !eligible {
				return s.record(tx, &result, &service, models.TxStatusRejected, models.ReasonNotEligible,
					&citizen.ID, nfcUID, userID, merchantName, 0)
			}
		}

		// (c) Layanan TANPA kuota (eligibility / log): cukup catat sukses.
		if service.Kind != models.ServiceKindQuota {
			return s.record(tx, &result, &service, models.TxStatusSuccess, "Layanan berhasil dicatat",
				&citizen.ID, nfcUID, userID, merchantName, 0)
		}

		// (d) Layanan ber-kuota: KUNCI baris kuota. Inilah inti pencegahan klaim ganda —
		//     jika KTP yang sama di-tap di dua mesin bersamaan, transaksi kedua MENUNGGU
		//     di sini sampai yang pertama commit, lalu membaca sisa terkini.
		var quota models.ServiceQuota
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("citizen_id = ? AND service_id = ? AND period = ?", citizen.ID, service.ID, period).
			First(&quota).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return s.record(tx, &result, &service, models.TxStatusRejected, models.ReasonNoQuotaConfig,
				&citizen.ID, nfcUID, userID, merchantName, 0)
		}
		if err != nil {
			return err
		}
		if quota.QuotaRemaining <= 0 {
			return s.record(tx, &result, &service, models.TxStatusRejected, models.ReasonQuotaEmpty,
				&citizen.ID, nfcUID, userID, merchantName, 0)
		}

		quota.QuotaRemaining--
		if err := tx.Save(&quota).Error; err != nil {
			return err
		}
		return s.record(tx, &result, &service, models.TxStatusSuccess, "Klaim berhasil",
			&citizen.ID, nfcUID, userID, merchantName, quota.QuotaRemaining)
	})

	if err != nil {
		return nil, err
	}
	return &result, nil
}

// isEligible menentukan kelayakan warga untuk satu layanan: pakai baris ServiceEligibility
// bila ada; bila tidak ada, jatuh ke Service.DefaultEligible (opt-out true / opt-in false).
func (s *ClaimService) isEligible(tx *gorm.DB, citizenID uuid.UUID, service *models.Service) (bool, error) {
	var elig models.ServiceEligibility
	err := tx.Where("citizen_id = ? AND service_id = ?", citizenID, service.ID).First(&elig).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return service.DefaultEligible, nil
	}
	if err != nil {
		return false, err
	}
	return elig.IsEligible, nil
}

// record menyimpan satu baris transaksi (sukses atau ditolak) dan mengisi result.
// Mengembalikan nil agar transaksi tetap di-commit (riwayat penolakan ikut tersimpan).
func (s *ClaimService) record(
	tx *gorm.DB, result *ClaimResult, service *models.Service, status, message string,
	citizenID *uuid.UUID, nfcUID string, userID uuid.UUID, merchantName string, quotaRemaining int,
) error {
	reason := ""
	if status == models.TxStatusRejected {
		reason = message
	}

	serviceID := service.ID
	trx := models.Transaction{
		CitizenID:    citizenID,
		NFCUID:       nfcUID,
		UserID:       userID,
		ServiceID:    &serviceID,
		ServiceCode:  service.Code,
		Status:       status,
		Reason:       reason,
		MerchantName: merchantName,
	}
	if err := tx.Create(&trx).Error; err != nil {
		return err
	}

	*result = ClaimResult{
		Status:         status,
		Message:        message,
		QuotaRemaining: quotaRemaining,
		TransactionID:  trx.ID,
	}
	return nil
}
