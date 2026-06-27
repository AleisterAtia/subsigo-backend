package models

import (
	"time"

	"github.com/google/uuid"
)

// ServiceEligibility adalah status kelayakan seorang warga untuk SATU layanan.
// Menggeneralisasi flag tunggal Citizen.IsEligible yang lama (yang berlaku global)
// menjadi per-(warga, layanan), sehingga warga bisa layak untuk subsidi tapi tidak
// untuk layanan lain — atau sebaliknya.
//
// Bila tidak ada baris untuk (warga, layanan), kelayakan jatuh ke Service.DefaultEligible.
type ServiceEligibility struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CitizenID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_eligibility_unique" json:"citizen_id"`
	ServiceID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_eligibility_unique" json:"service_id"`
	IsEligible bool      `gorm:"not null;default:true" json:"is_eligible"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName memaksa nama tabel agar deterministik & konsisten.
func (ServiceEligibility) TableName() string { return "service_eligibilities" }
