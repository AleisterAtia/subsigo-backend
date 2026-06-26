package services

import (
	"errors"

	"gorm.io/gorm"

	"github.com/sitepat/subsigo-backend/internal/models"
	"github.com/sitepat/subsigo-backend/internal/repositories"
	"github.com/sitepat/subsigo-backend/pkg/hash"
	"github.com/sitepat/subsigo-backend/pkg/token"
)

// ErrInvalidCredentials dikembalikan saat username/password salah.
var ErrInvalidCredentials = errors.New("username atau password salah")

// ErrAccountDisabled dikembalikan saat akun ada & kredensial benar tapi dinonaktifkan.
var ErrAccountDisabled = errors.New("akun dinonaktifkan, hubungi admin")

// AuthService menangani logika autentikasi.
type AuthService struct {
	users  *repositories.UserRepository
	tokens *token.Manager
}

func NewAuthService(users *repositories.UserRepository, tokens *token.Manager) *AuthService {
	return &AuthService{users: users, tokens: tokens}
}

// Login memvalidasi kredensial dan mengembalikan JWT beserta data user.
func (s *AuthService) Login(username, password string) (string, *models.User, error) {
	u, err := s.users.FindByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, err
	}

	if !hash.Check(password, u.PasswordHash) {
		return "", nil, ErrInvalidCredentials
	}

	// Cek setelah verifikasi password agar tidak membocorkan username mana yang nonaktif.
	if !u.IsActive {
		return "", nil, ErrAccountDisabled
	}

	tok, err := s.tokens.Generate(u.ID, u.Role, u.MerchantName)
	if err != nil {
		return "", nil, err
	}
	return tok, u, nil
}
