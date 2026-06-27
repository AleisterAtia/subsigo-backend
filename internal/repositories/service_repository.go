package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sitepat/subsigo-backend/internal/models"
)

// ServiceRepository menangani akses data tabel services (katalog layanan).
type ServiceRepository struct {
	db *gorm.DB
}

func NewServiceRepository(db *gorm.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

// FindByCode mencari layanan berdasarkan Code unik (mis. "LPG_3KG").
// Mengembalikan gorm.ErrRecordNotFound bila tidak ada.
func (r *ServiceRepository) FindByCode(code string) (*models.Service, error) {
	var s models.Service
	if err := r.db.Where("code = ?", code).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

// FindByID mencari layanan berdasarkan ID. gorm.ErrRecordNotFound bila tidak ada.
func (r *ServiceRepository) FindByID(id uuid.UUID) (*models.Service, error) {
	var s models.Service
	if err := r.db.First(&s, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

// Update menyimpan perubahan satu layanan. Hanya kolom yang ditentukan pemanggil
// (lihat ServiceService.Update) yang diubah; Code sengaja TIDAK pernah diubah agar
// denormalisasi service_code di service_quotas/transactions tetap konsisten.
func (r *ServiceRepository) Update(s *models.Service) error {
	return r.db.Save(s).Error
}

// List mengembalikan daftar layanan (diurutkan per Code). Bila activeOnly true,
// hanya layanan yang IsActive.
func (r *ServiceRepository) List(activeOnly bool) ([]models.Service, error) {
	q := r.db.Model(&models.Service{})
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	var services []models.Service
	if err := q.Order("code ASC").Find(&services).Error; err != nil {
		return nil, err
	}
	return services, nil
}

func (r *ServiceRepository) Create(s *models.Service) error {
	return r.db.Create(s).Error
}
