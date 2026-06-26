package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sitepat/subsigo-backend/internal/models"
)

// UserRepository menangani akses data tabel users.
type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByUsername mencari user berdasarkan username.
// Mengembalikan gorm.ErrRecordNotFound bila tidak ditemukan.
func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	var u models.User
	if err := r.db.Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByID mencari user berdasarkan ID.
func (r *UserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var u models.User
	if err := r.db.First(&u, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// Create menyimpan user baru.
func (r *UserRepository) Create(u *models.User) error {
	return r.db.Create(u).Error
}

// List mengembalikan daftar user (terbaru lebih dulu) beserta totalnya untuk pagination.
// Bila search tidak kosong, mencocokkan username / merchant_name (ILIKE).
func (r *UserRepository) List(search string, limit, offset int) ([]models.User, int64, error) {
	q := r.db.Model(&models.User{})
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("username ILIKE ? OR merchant_name ILIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []models.User
	err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// UpdateFields memperbarui kolom tertentu user berdasarkan id. Mengembalikan jumlah
// baris terpengaruh (0 berarti user tidak ditemukan). Memakai map agar hanya kolom
// yang diberikan yang di-update (termasuk zero-value seperti is_active=false).
func (r *UserRepository) UpdateFields(id uuid.UUID, fields map[string]any) (int64, error) {
	res := r.db.Model(&models.User{}).Where("id = ?", id).Updates(fields)
	return res.RowsAffected, res.Error
}
