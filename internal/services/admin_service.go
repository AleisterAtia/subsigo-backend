package services

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sitepat/subsigo-backend/internal/models"
	"github.com/sitepat/subsigo-backend/internal/repositories"
)

// ErrCitizenNotFound dikembalikan saat warga tidak ditemukan.
var ErrCitizenNotFound = errors.New("warga tidak ditemukan")

// AdminService menangani operasi admin: registrasi warga, kelayakan, kuota, monitoring.
type AdminService struct {
	citizens     *repositories.CitizenRepository
	quotas       *repositories.QuotaRepository
	transactions *repositories.TransactionRepository
}

func NewAdminService(
	citizens *repositories.CitizenRepository,
	quotas *repositories.QuotaRepository,
	transactions *repositories.TransactionRepository,
) *AdminService {
	return &AdminService{citizens: citizens, quotas: quotas, transactions: transactions}
}

// ListCitizens mengembalikan daftar warga (dengan pencarian opsional) + total untuk pagination.
func (s *AdminService) ListCitizens(search string, limit, offset int) ([]models.Citizen, int64, error) {
	return s.citizens.List(search, limit, offset)
}

// GetCitizen mengembalikan satu warga beserta kuotanya. ErrCitizenNotFound bila tidak ada.
func (s *AdminService) GetCitizen(id uuid.UUID) (*models.Citizen, error) {
	c, err := s.citizens.FindByIDWithQuotas(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCitizenNotFound
		}
		return nil, err
	}
	return c, nil
}

// RegisterCitizen mendaftarkan warga baru (default langsung layak/eligible).
func (s *AdminService) RegisterCitizen(nik, nfcUID, name string) (*models.Citizen, error) {
	c := &models.Citizen{
		NIK:        nik,
		NFCUID:     nfcUID,
		Name:       name,
		IsEligible: true,
	}
	if err := s.citizens.Create(c); err != nil {
		return nil, err
	}
	return c, nil
}

// SetEligibility mengubah status kelayakan subsidi warga.
func (s *AdminService) SetEligibility(id uuid.UUID, eligible bool) error {
	rows, err := s.citizens.SetEligibility(id, eligible)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCitizenNotFound
	}
	return nil
}

// SetQuota menetapkan/mereset kuota warga untuk satu komoditas pada satu periode.
func (s *AdminService) SetQuota(citizenID uuid.UUID, commodity, period string, total int) (*models.SubsidyQuota, error) {
	if _, err := s.citizens.FindByID(citizenID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCitizenNotFound
		}
		return nil, err
	}
	return s.quotas.Upsert(citizenID, commodity, period, total)
}

// ListTransactions mengembalikan riwayat transaksi sesuai filter untuk monitoring,
// beserta total baris yang cocok (untuk pagination).
func (s *AdminService) ListTransactions(f repositories.TransactionFilter) ([]models.Transaction, int64, error) {
	return s.transactions.List(f)
}
