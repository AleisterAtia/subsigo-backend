package models

import (
	"time"

	"github.com/google/uuid"
)

// SubsidyQuota adalah jatah subsidi seorang warga untuk satu komoditas pada satu periode.
// Baris inilah yang dikunci (SELECT ... FOR UPDATE) saat klaim untuk mencegah race condition.
//
// Unik per (citizen_id, commodity, period) sehingga kuota bisa di-reset per periode
// (mis. bulanan: period = "2026-06") tanpa menimpa riwayat periode sebelumnya.
type SubsidyQuota struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CitizenID      uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_quota_unique" json:"citizen_id"`
	Commodity      string    `gorm:"type:varchar(32);not null;uniqueIndex:idx_quota_unique" json:"commodity"`
	Period         string    `gorm:"type:varchar(7);not null;uniqueIndex:idx_quota_unique" json:"period"` // format "YYYY-MM"
	QuotaTotal     int       `gorm:"not null;default:0" json:"quota_total"`
	QuotaRemaining int       `gorm:"not null;default:0" json:"quota_remaining"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName memaksa nama tabel jamak agar konsisten dengan model lain
// (GORM default tidak menjamakkan "quota").
func (SubsidyQuota) TableName() string { return "subsidy_quotas" }
