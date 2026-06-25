package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sitepat/subsigo-backend/internal/models"
)

// CitizenRepository menangani akses data tabel citizens.
type CitizenRepository struct {
	db *gorm.DB
}

func NewCitizenRepository(db *gorm.DB) *CitizenRepository {
	return &CitizenRepository{db: db}
}

func (r *CitizenRepository) Create(c *models.Citizen) error {
	return r.db.Create(c).Error
}

func (r *CitizenRepository) FindByID(id uuid.UUID) (*models.Citizen, error) {
	var c models.Citizen
	if err := r.db.First(&c, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

// SetEligibility memperbarui status kelayakan. Mengembalikan jumlah baris yang terpengaruh
// (0 berarti warga tidak ditemukan).
func (r *CitizenRepository) SetEligibility(id uuid.UUID, eligible bool) (int64, error) {
	res := r.db.Model(&models.Citizen{}).Where("id = ?", id).Update("is_eligible", eligible)
	return res.RowsAffected, res.Error
}
