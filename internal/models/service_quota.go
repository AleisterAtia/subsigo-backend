package models

import (
	"time"

	"github.com/google/uuid"
)

// ServiceQuota adalah jatah seorang warga untuk satu layanan ber-Kind "quota" pada
// satu periode. Baris inilah yang dikunci (SELECT ... FOR UPDATE) saat klaim untuk
// mencegah race condition (klaim ganda).
//
// Generalisasi dari SubsidyQuota lama: kolom Commodity diganti ServiceID. Unik per
// (citizen_id, service_id, period) sehingga kuota bisa di-reset per periode
// (mis. bulanan: period = "2026-06") tanpa menimpa riwayat periode sebelumnya.
//
// Catatan: nama index unik sengaja "idx_service_quota_unique" (BUKAN "idx_quota_unique"
// milik tabel subsidy_quotas lama) — di PostgreSQL nama index unik per schema, jadi
// memakai nama lama akan bentrok selama tabel lama masih ada.
type ServiceQuota struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CitizenID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_service_quota_unique" json:"citizen_id"`
	ServiceID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_service_quota_unique" json:"service_id"`
	// ServiceCode: denormalisasi Code layanan agar dashboard bisa memberi label tanpa join.
	ServiceCode    string    `gorm:"type:varchar(32);not null" json:"service_code"`
	Period         string    `gorm:"type:varchar(7);not null;uniqueIndex:idx_service_quota_unique" json:"period"` // "YYYY-MM"
	QuotaTotal     int       `gorm:"not null;default:0" json:"quota_total"`
	QuotaRemaining int       `gorm:"not null;default:0" json:"quota_remaining"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName memaksa nama tabel jamak agar konsisten dengan model lain.
func (ServiceQuota) TableName() string { return "service_quotas" }
