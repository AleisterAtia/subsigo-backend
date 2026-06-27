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

// FindByIDWithQuotas mengambil satu warga beserta seluruh kuotanya
// (diurutkan periode terbaru lebih dulu) untuk halaman detail di dashboard.
func (r *CitizenRepository) FindByIDWithQuotas(id uuid.UUID) (*models.Citizen, error) {
	var c models.Citizen
	err := r.db.
		Preload("Quotas", func(db *gorm.DB) *gorm.DB {
			return db.Order("period DESC, service_code ASC")
		}).
		First(&c, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// List mengembalikan daftar warga (terbaru lebih dulu) beserta total seluruh data
// (untuk pagination). Bila search tidak kosong, mencocokkan NIK / NFC UID / nama
// secara case-insensitive (ILIKE).
func (r *CitizenRepository) List(search string, limit, offset int) ([]models.Citizen, int64, error) {
	q := r.db.Model(&models.Citizen{})
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("nik ILIKE ? OR nfc_uid ILIKE ? OR name ILIKE ?", like, like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var citizens []models.Citizen
	err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&citizens).Error
	if err != nil {
		return nil, 0, err
	}
	return citizens, total, nil
}
