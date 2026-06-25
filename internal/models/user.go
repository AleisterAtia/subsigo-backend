package models

import (
	"time"

	"github.com/google/uuid"
)

// User adalah akun yang login ke sistem: petugas lapangan (merchant) atau admin kelurahan.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Username     string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	Role         string    `gorm:"type:varchar(16);not null;default:merchant" json:"role"`
	// MerchantName: nama lokasi SPBU/pangkalan, hanya relevan untuk role merchant.
	MerchantName string    `gorm:"type:varchar(128)" json:"merchant_name,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
