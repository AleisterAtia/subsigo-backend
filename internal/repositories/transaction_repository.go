package repositories

import (
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

// List mengembalikan transaksi terbaru lebih dulu (untuk monitoring admin).
func (r *TransactionRepository) List(limit int) ([]models.Transaction, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var txs []models.Transaction
	err := r.db.Order("created_at DESC").Limit(limit).Find(&txs).Error
	return txs, err
}
