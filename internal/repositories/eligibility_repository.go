package repositories

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/sitepat/subsigo-backend/internal/models"
)

// EligibilityRepository menangani akses data tabel service_eligibilities
// (kelayakan warga per-layanan).
type EligibilityRepository struct {
	db *gorm.DB
}

func NewEligibilityRepository(db *gorm.DB) *EligibilityRepository {
	return &EligibilityRepository{db: db}
}

// FindByCitizenService mengembalikan baris kelayakan untuk (warga, layanan) bila ada.
// Mengembalikan (nil, nil) bila TIDAK ada — pemanggil yang memutuskan default lewat
// Service.DefaultEligible (lihat ClaimService).
func (r *EligibilityRepository) FindByCitizenService(citizenID, serviceID uuid.UUID) (*models.ServiceEligibility, error) {
	var e models.ServiceEligibility
	err := r.db.Where("citizen_id = ? AND service_id = ?", citizenID, serviceID).First(&e).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// Upsert menetapkan kelayakan (warga, layanan) secara ATOMIK lewat INSERT ... ON CONFLICT,
// meniru pola QuotaRepository.Upsert agar aman dari race dua admin yang menyetel bersamaan.
func (r *EligibilityRepository) Upsert(citizenID, serviceID uuid.UUID, eligible bool) (*models.ServiceEligibility, error) {
	e := models.ServiceEligibility{
		CitizenID:  citizenID,
		ServiceID:  serviceID,
		IsEligible: eligible,
	}
	err := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "citizen_id"}, {Name: "service_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"is_eligible": eligible,
			"updated_at":  time.Now().UTC(),
		}),
	}).Create(&e).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}
