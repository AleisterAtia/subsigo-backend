package repositories

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sitepat/subsigo-backend/internal/models"
)

// TransactionRepository menangani akses data tabel transactions.
type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// TransactionFilter menampung kriteria filter & pagination untuk monitoring admin.
// Field kosong/nil berarti "tidak difilter".
type TransactionFilter struct {
	Status       string     // "success" | "rejected"
	ServiceCode  string     // kode layanan (mis. "LPG_3KG"); query lama ?commodity= dipetakan ke sini
	MerchantName string     // pencarian ILIKE
	UserID       *uuid.UUID // petugas tertentu
	From         *time.Time // created_at >= From (inklusif)
	To           *time.Time // created_at < To (eksklusif)
	Limit        int
	Offset       int
}

// List mengembalikan transaksi sesuai filter (terbaru lebih dulu) beserta total
// baris yang cocok (untuk pagination).
func (r *TransactionRepository) List(f TransactionFilter) ([]models.Transaction, int64, error) {
	q := r.db.Model(&models.Transaction{})

	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.ServiceCode != "" {
		q = q.Where("service_code = ?", f.ServiceCode)
	}
	if f.MerchantName != "" {
		q = q.Where("merchant_name ILIKE ?", "%"+f.MerchantName+"%")
	}
	if f.UserID != nil {
		q = q.Where("user_id = ?", *f.UserID)
	}
	if f.From != nil {
		q = q.Where("created_at >= ?", *f.From)
	}
	if f.To != nil {
		q = q.Where("created_at < ?", *f.To)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var txs []models.Transaction
	err := q.Order("created_at DESC").Limit(f.Limit).Offset(f.Offset).Find(&txs).Error
	if err != nil {
		return nil, 0, err
	}
	return txs, total, nil
}
