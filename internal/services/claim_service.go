package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/sitepat/subsigo-backend/internal/models"
)

// ErrInvalidCommodity dikembalikan saat jenis komoditas tidak dikenal.
var ErrInvalidCommodity = errors.New("jenis komoditas tidak valid")

// ClaimResult adalah hasil pemrosesan klaim yang dikirim ke aplikasi petugas.
type ClaimResult struct {
	Status         string    `json:"status"`          // "success" | "rejected"
	Message        string    `json:"message"`         // pesan/alasan untuk ditampilkan
	QuotaRemaining int       `json:"quota_remaining"` // sisa kuota setelah klaim (jika sukses)
	TransactionID  uuid.UUID `json:"transaction_id"`
}

// ClaimService menangani logika klaim subsidi.
// Menyimpan *gorm.DB karena alur kritis dijalankan dalam satu transaksi dengan row-locking.
type ClaimService struct {
	db *gorm.DB
}

func NewClaimService(db *gorm.DB) *ClaimService {
	return &ClaimService{db: db}
}

// Claim memproses satu klaim subsidi secara atomik & aman terhadap race condition.
//
// Alur:
//  1. Validasi komoditas.
//  2. Dalam SATU transaksi:
//     a. Cari warga berdasarkan NFC UID.
//     b. Pastikan warga layak (eligible).
//     c. Kunci baris kuota (SELECT ... FOR UPDATE) untuk periode berjalan.
//     d. Pastikan sisa kuota > 0.
//     e. Potong kuota & catat transaksi sukses.
//
// Setiap penolakan tetap dicatat sebagai transaksi (sesuai PRD: catat gagal & sukses).
func (s *ClaimService) Claim(userID uuid.UUID, merchantName, nfcUID, commodity string) (*ClaimResult, error) {
	if !models.IsValidCommodity(commodity) {
		return nil, ErrInvalidCommodity
	}

	// Periode kuota saat ini, format "YYYY-MM" (kuota di-reset per bulan).
	period := time.Now().UTC().Format("2006-01")

	var result ClaimResult

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// (a) Cari warga berdasarkan UID e-KTP.
		var citizen models.Citizen
		err := tx.Where("nfc_uid = ?", nfcUID).First(&citizen).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return s.record(tx, &result, models.TxStatusRejected, models.ReasonNotRegistered,
				nil, nfcUID, userID, merchantName, commodity, 0)
		}
		if err != nil {
			return err
		}

		// (b) Cek kelayakan subsidi.
		if !citizen.IsEligible {
			return s.record(tx, &result, models.TxStatusRejected, models.ReasonNotEligible,
				&citizen.ID, nfcUID, userID, merchantName, commodity, 0)
		}

		// (c) KUNCI baris kuota. Inilah inti pencegahan klaim ganda: jika KTP yang sama
		//     di-tap di dua mesin pada saat bersamaan, transaksi kedua akan MENUNGGU
		//     di sini sampai transaksi pertama selesai (commit), lalu membaca sisa terkini.
		var quota models.SubsidyQuota
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("citizen_id = ? AND commodity = ? AND period = ?", citizen.ID, commodity, period).
			First(&quota).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return s.record(tx, &result, models.TxStatusRejected, models.ReasonNoQuotaConfig,
				&citizen.ID, nfcUID, userID, merchantName, commodity, 0)
		}
		if err != nil {
			return err
		}

		// (d) Cek sisa kuota.
		if quota.QuotaRemaining <= 0 {
			return s.record(tx, &result, models.TxStatusRejected, models.ReasonQuotaEmpty,
				&citizen.ID, nfcUID, userID, merchantName, commodity, 0)
		}

		// (e) Potong kuota & catat transaksi sukses.
		quota.QuotaRemaining--
		if err := tx.Save(&quota).Error; err != nil {
			return err
		}
		return s.record(tx, &result, models.TxStatusSuccess, "Klaim berhasil",
			&citizen.ID, nfcUID, userID, merchantName, commodity, quota.QuotaRemaining)
	})

	if err != nil {
		return nil, err
	}
	return &result, nil
}

// record menyimpan satu baris transaksi (sukses atau ditolak) dan mengisi result.
// Mengembalikan nil agar transaksi tetap di-commit (riwayat penolakan ikut tersimpan).
func (s *ClaimService) record(
	tx *gorm.DB, result *ClaimResult, status, message string,
	citizenID *uuid.UUID, nfcUID string, userID uuid.UUID,
	merchantName, commodity string, quotaRemaining int,
) error {
	reason := ""
	if status == models.TxStatusRejected {
		reason = message
	}

	trx := models.Transaction{
		CitizenID:    citizenID,
		NFCUID:       nfcUID,
		UserID:       userID,
		Commodity:    commodity,
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
