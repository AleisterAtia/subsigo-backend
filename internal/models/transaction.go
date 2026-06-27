package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Transaction mencatat setiap upaya menjalankan layanan, baik berhasil maupun ditolak
// (sesuai PRD: "Mencatat setiap riwayat yang berhasil maupun gagal").
//
// Dulu khusus klaim subsidi (kolom Commodity); kini generik per-layanan: ServiceID +
// ServiceCode menggantikan Commodity. ServiceCode didenormalisasi agar riwayat tetap
// terbaca meski layanan kelak di-rename/nonaktifkan (baris audit harus imutabel —
// karena itu service_id sengaja TANPA foreign key).
type Transaction struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	// CitizenID nullable: jika UID yang di-tap tidak terdaftar, transaksi tetap dicatat.
	CitizenID *uuid.UUID `gorm:"type:uuid;index" json:"citizen_id,omitempty"`
	// NFCUID selalu dicatat apa adanya, walau tidak ketemu di tabel citizens.
	NFCUID string    `gorm:"type:varchar(64);not null;index" json:"nfc_uid"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"` // petugas yang melayani
	// ServiceID & ServiceCode: layanan yang dijalankan. Nullable agar AutoMigrate aman
	// menambah kolom ini ke tabel transactions yang sudah berisi data (lihat cmd/migrate).
	ServiceID    *uuid.UUID `gorm:"type:uuid;index" json:"service_id,omitempty"`
	ServiceCode  string     `gorm:"type:varchar(32);index" json:"service_code,omitempty"`
	Status       string     `gorm:"type:varchar(16);not null" json:"status"` // success | rejected
	Reason       string     `gorm:"type:varchar(128)" json:"reason,omitempty"`
	MerchantName string     `gorm:"type:varchar(128)" json:"merchant_name,omitempty"`
	// Metadata: data tambahan khas tiap layanan (mis. nama poliklinik, jenis bantuan).
	// Memakai json.RawMessage (alias []byte) + kolom jsonb — TANPA dependensi baru,
	// ter-serialize sebagai JSON mentah (bukan base64). Nullable (tanpa NOT NULL/default)
	// agar AutoMigrate aman pada tabel yang sudah berisi data.
	Metadata  json.RawMessage `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt time.Time       `gorm:"index" json:"created_at"` // index untuk monitoring (ORDER BY created_at DESC)
}
