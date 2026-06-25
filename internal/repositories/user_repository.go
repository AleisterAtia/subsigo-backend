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
