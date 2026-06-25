package repositories

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sitepat/subsigo-backend/internal/models"
)

// QuotaRepository menangani akses data tabel subsidy_quotas.
type QuotaRepository struct {
	db *gorm.DB
}

func NewQuotaRepository(db *gorm.DB) *QuotaRepository {
	return &QuotaRepository{db: db}
}

// Upsert membuat atau memperbarui kuota untuk (citizen, commodity, period).
// Menetapkan total dan mereset sisa kuota = total (inisialisasi ulang).
func (r *QuotaRepository) Upsert(citizenID uuid.UUID, commodity, period string, total int) (*models.SubsidyQuota, error) {
	var q models.SubsidyQuota
	err := r.db.Where("citizen_id = ? AND commodity = ? AND period = ?", citizenID, commodity, period).First(&q).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		q = models.SubsidyQuota{
			CitizenID:      citizenID,
			Commodity:      commodity,
			Period:         period,
			QuotaTotal:     total,
			QuotaRemaining: total,
		}
		if err := r.db.Create(&q).Error; err != nil {
			return nil, err
		}
		return &q, nil
	}
	if err != nil {
		return nil, err
	}

	q.QuotaTotal = total
	q.QuotaRemaining = total
	if err := r.db.Save(&q).Error; err != nil {
		return nil, err
	}
	return &q, nil
}
