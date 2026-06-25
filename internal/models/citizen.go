package models

import (
	"time"

	"github.com/google/uuid"
)

// Citizen adalah data kependudukan warga (replika/simulasi, belum terhubung Dukcapil).
// NFCUID adalah kunci pencarian utama saat e-KTP di-tap oleh petugas.
type Citizen struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	NIK        string    `gorm:"type:varchar(16);uniqueIndex;not null" json:"nik"`
	NFCUID     string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"nfc_uid"`
	Name       string    `gorm:"type:varchar(128);not null" json:"name"`
	// IsEligible: status kelayakan subsidi yang diatur admin (Aktif/Tidak Aktif).
	IsEligible bool      `gorm:"not null;default:true" json:"is_eligible"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	Quotas []SubsidyQuota `gorm:"foreignKey:CitizenID" json:"quotas,omitempty"`
}
