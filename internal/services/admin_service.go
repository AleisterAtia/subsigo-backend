package services

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sitepat/subsigo-backend/internal/models"
	"github.com/sitepat/subsigo-backend/internal/repositories"
)

var (
	// ErrCitizenNotFound dikembalikan saat warga tidak ditemukan.
	ErrCitizenNotFound = errors.New("warga tidak ditemukan")
	// ErrServiceNotFound dikembalikan saat kode layanan tidak dikenal.
	ErrServiceNotFound = errors.New("layanan tidak ditemukan")
	// ErrServiceNotQuota dikembalikan saat admin mencoba menyetel kuota untuk layanan
	// yang bukan ber-Kind "quota" (kuota tak akan pernah dipakai alur klaimnya).
	ErrServiceNotQuota = errors.New("layanan ini tidak memakai kuota")
)

// AdminService menangani operasi admin: registrasi warga, kelayakan, kuota, monitoring.
type AdminService struct {
	citizens     *repositories.CitizenRepository
	services     *repositories.ServiceRepository
	quotas       *repositories.QuotaRepository
	eligibility  *repositories.EligibilityRepository
	transactions *repositories.TransactionRepository
}

func NewAdminService(
	citizens *repositories.CitizenRepository,
	services *repositories.ServiceRepository,
	quotas *repositories.QuotaRepository,
	eligibility *repositories.EligibilityRepository,
	transactions *repositories.TransactionRepository,
) *AdminService {
	return &AdminService{
		citizens:     citizens,
		services:     services,
		quotas:       quotas,
		eligibility:  eligibility,
		transactions: transactions,
	}
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

// RegisterCitizen mendaftarkan warga baru. Kelayakan TIDAK lagi di-set di sini —
// kini per-layanan (lihat SetEligibility) dengan default dari Service.DefaultEligible.
func (s *AdminService) RegisterCitizen(nik, nfcUID, name string) (*models.Citizen, error) {
	c := &models.Citizen{
		NIK:    nik,
		NFCUID: nfcUID,
		Name:   name,
	}
	if err := s.citizens.Create(c); err != nil {
		return nil, err
	}
	return c, nil
}

// SetEligibility menetapkan kelayakan warga. Bila serviceCode KOSONG, diterapkan ke
// SEMUA layanan aktif yang membutuhkan kelayakan (quota/eligibility) — meniru tombol
// kelayakan global lama. Bila diisi, hanya untuk layanan tersebut.
func (s *AdminService) SetEligibility(citizenID uuid.UUID, serviceCode string, eligible bool) error {
	if _, err := s.citizens.FindByID(citizenID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCitizenNotFound
		}
		return err
	}

	targets, err := s.eligibilityTargets(serviceCode)
	if err != nil {
		return err
	}
	for _, sv := range targets {
		if _, err := s.eligibility.Upsert(citizenID, sv.ID, eligible); err != nil {
			return err
		}
	}
	return nil
}

// eligibilityTargets memilih layanan sasaran SetEligibility: satu layanan (bila code diisi)
// atau seluruh layanan aktif yang butuh kelayakan (bila code kosong).
func (s *AdminService) eligibilityTargets(serviceCode string) ([]models.Service, error) {
	if serviceCode != "" {
		sv, err := s.services.FindByCode(serviceCode)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrServiceNotFound
			}
			return nil, err
		}
		return []models.Service{*sv}, nil
	}

	all, err := s.services.List(true)
	if err != nil {
		return nil, err
	}
	var targets []models.Service
	for _, sv := range all {
		if sv.Kind == models.ServiceKindQuota || sv.Kind == models.ServiceKindEligibility {
			targets = append(targets, sv)
		}
	}
	return targets, nil
}

// SetQuota menetapkan/mereset kuota warga untuk satu layanan ber-Kind "quota" pada satu periode.
func (s *AdminService) SetQuota(citizenID uuid.UUID, serviceCode, period string, total int) (*models.ServiceQuota, error) {
	if _, err := s.citizens.FindByID(citizenID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCitizenNotFound
		}
		return nil, err
	}

	sv, err := s.services.FindByCode(serviceCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrServiceNotFound
		}
		return nil, err
	}
	if sv.Kind != models.ServiceKindQuota {
		return nil, ErrServiceNotQuota
	}
	return s.quotas.Upsert(citizenID, sv.ID, sv.Code, period, total)
}

// ListTransactions mengembalikan riwayat transaksi sesuai filter untuk monitoring,
// beserta total baris yang cocok (untuk pagination).
func (s *AdminService) ListTransactions(f repositories.TransactionFilter) ([]models.Transaction, int64, error) {
	return s.transactions.List(f)
}
