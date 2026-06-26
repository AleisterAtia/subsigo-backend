package services

import (
	"errors"

	"github.com/google/uuid"

	"github.com/sitepat/subsigo-backend/internal/models"
	"github.com/sitepat/subsigo-backend/internal/repositories"
	"github.com/sitepat/subsigo-backend/pkg/hash"
)

// ErrUserNotFound dikembalikan saat user tidak ditemukan.
var ErrUserNotFound = errors.New("user tidak ditemukan")

// ErrNoUpdateFields dikembalikan saat request update tidak membawa field apa pun.
var ErrNoUpdateFields = errors.New("tidak ada field yang diperbarui")

// UserService menangani manajemen akun (admin & petugas) oleh admin.
type UserService struct {
	users *repositories.UserRepository
}

func NewUserService(users *repositories.UserRepository) *UserService {
	return &UserService{users: users}
}

// CreateUser membuat akun baru dengan password yang di-hash bcrypt.
// Duplikasi username diteruskan sebagai gorm.ErrDuplicatedKey untuk ditangani handler.
func (s *UserService) CreateUser(username, password, role, merchantName string) (*models.User, error) {
	hashed, err := hash.Hash(password)
	if err != nil {
		return nil, err
	}
	u := &models.User{
		Username:     username,
		PasswordHash: hashed,
		Role:         role,
		MerchantName: merchantName,
		IsActive:     true,
	}
	if err := s.users.Create(u); err != nil {
		return nil, err
	}
	return u, nil
}

// ListUsers mengembalikan daftar user (pencarian opsional) + total untuk pagination.
func (s *UserService) ListUsers(search string, limit, offset int) ([]models.User, int64, error) {
	return s.users.List(search, limit, offset)
}

// UpdateUser memperbarui field user yang diberikan (semua opsional). Password, bila
// ada, akan di-hash ulang. Mengembalikan user terbaru, atau ErrUserNotFound bila id
// tidak ada, atau ErrNoUpdateFields bila tidak ada field sama sekali.
func (s *UserService) UpdateUser(id uuid.UUID, role, merchantName, password *string, isActive *bool) (*models.User, error) {
	fields := map[string]any{}
	if role != nil {
		fields["role"] = *role
	}
	if merchantName != nil {
		fields["merchant_name"] = *merchantName
	}
	if isActive != nil {
		fields["is_active"] = *isActive
	}
	if password != nil {
		hashed, err := hash.Hash(*password)
		if err != nil {
			return nil, err
		}
		fields["password_hash"] = hashed
	}
	if len(fields) == 0 {
		return nil, ErrNoUpdateFields
	}

	rows, err := s.users.UpdateFields(id, fields)
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, ErrUserNotFound
	}
	return s.users.FindByID(id)
}
