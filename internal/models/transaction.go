package models

import (
	"time"

	"github.com/google/uuid"
)

// Transaction mencatat setiap upaya klaim subsidi, baik yang berhasil maupun ditolak
// (sesuai PRD: "Mencatat setiap riwayat klaim yang berhasil maupun gagal").
type Transaction struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	// CitizenID nullable: jika UID yang di-tap tidak terdaftar, transaksi tetap dicatat.
	CitizenID *uuid.UUID `gorm:"type:uuid;index" json:"citizen_id,omitempty"`
	// NFCUID selalu dicatat apa adanya, walau tidak ketemu di tabel citizens.
	NFCUID       string    `gorm:"type:varchar(64);not null;index" json:"nfc_uid"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"` // petugas yang melayani
	Commodity    string    `gorm:"type:varchar(32);not null" json:"commodity"`
	Status       string    `gorm:"type:varchar(16);not null" json:"status"` // success | rejected
	Reason       string    `gorm:"type:varchar(128)" json:"reason,omitempty"`
	MerchantName string    `gorm:"type:varchar(128)" json:"merchant_name,omitempty"`
	CreatedAt    time.Time `gorm:"index" json:"created_at"` // index untuk monitoring (ORDER BY created_at DESC)
}
