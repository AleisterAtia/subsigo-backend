package repositories

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/sitepat/subsigo-backend/internal/models"
)

// QuotaRepository menangani akses data tabel service_quotas.
type QuotaRepository struct {
	db *gorm.DB
}

func NewQuotaRepository(db *gorm.DB) *QuotaRepository {
	return &QuotaRepository{db: db}
}

// Upsert membuat atau memperbarui kuota untuk (citizen, service, period) secara
// ATOMIK lewat INSERT ... ON CONFLICT. Menetapkan total dan mereset sisa kuota = total.
// serviceCode didenormalisasi ikut disimpan agar dashboard bisa memberi label tanpa join.
//
// Pola lama (SELECT lalu Create/Save) tidak aman bila dua admin menyetel kuota yang
// sama bersamaan — keduanya bisa lolos SELECT lalu satu Create kena unique-violation
// pada idx_service_quota_unique. ON CONFLICT menghilangkan race itu di tingkat database.
func (r *QuotaRepository) Upsert(citizenID, serviceID uuid.UUID, serviceCode, period string, total int) (*models.ServiceQuota, error) {
	q := models.ServiceQuota{
		CitizenID:      citizenID,
		ServiceID:      serviceID,
		ServiceCode:    serviceCode,
		Period:         period,
		QuotaTotal:     total,
		QuotaRemaining: total,
	}
	err := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "citizen_id"}, {Name: "service_id"}, {Name: "period"}},
		DoUpdates: clause.Assignments(map[string]any{
			"quota_total":     total,
			"quota_remaining": total,
			"service_code":    serviceCode,
			"updated_at":      time.Now().UTC(),
		}),
	}).Create(&q).Error
	if err != nil {
		return nil, err
	}

	// Muat ulang baris kanonik agar timestamp akurat di response, baik pada jalur
	// insert maupun update (RETURNING pada conflict tidak selalu mengisi semua kolom).
	var saved models.ServiceQuota
	err = r.db.
		Where("citizen_id = ? AND service_id = ? AND period = ?", citizenID, serviceID, period).
		First(&saved).Error
	if err != nil {
		return nil, err
	}
	return &saved, nil
}
