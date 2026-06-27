package services

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sitepat/subsigo-backend/internal/models"
	"github.com/sitepat/subsigo-backend/internal/repositories"
)

// ServiceService menangani manajemen katalog layanan (admin).
type ServiceService struct {
	services *repositories.ServiceRepository
}

func NewServiceService(services *repositories.ServiceRepository) *ServiceService {
	return &ServiceService{services: services}
}

// List mengembalikan seluruh layanan (aktif & nonaktif) untuk dashboard admin.
func (s *ServiceService) List() ([]models.Service, error) {
	return s.services.List(false)
}

// Create membuat layanan baru (default IsActive=true). Mengembalikan gorm.ErrDuplicatedKey
// bila code sudah dipakai (dipetakan handler ke 409). Validasi format ditangani handler.
func (s *ServiceService) Create(code, name, kind string, defaultEligible bool) (*models.Service, error) {
	svc := &models.Service{
		Code:            code,
		Name:            name,
		Kind:            kind,
		DefaultEligible: defaultEligible,
		IsActive:        true,
	}
	if err := s.services.Create(svc); err != nil {
		return nil, err
	}
	return svc, nil
}

// Update memperbarui field yang diberikan (nil = tidak diubah). Code TIDAK bisa diubah,
// demi menjaga konsistensi denormalisasi service_code di service_quotas & transactions.
func (s *ServiceService) Update(id uuid.UUID, name, kind *string, defaultEligible, isActive *bool) (*models.Service, error) {
	svc, err := s.services.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrServiceNotFound
		}
		return nil, err
	}
	if name != nil {
		svc.Name = *name
	}
	if kind != nil {
		svc.Kind = *kind
	}
	if defaultEligible != nil {
		svc.DefaultEligible = *defaultEligible
	}
	if isActive != nil {
		svc.IsActive = *isActive
	}
	if err := s.services.Update(svc); err != nil {
		return nil, err
	}
	return svc, nil
}
