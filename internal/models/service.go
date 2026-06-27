package models

import (
	"time"

	"github.com/google/uuid"
)

// Service adalah satu layanan yang bisa dijalankan lewat tap e-KTP.
// Menggantikan konsep "komoditas" yang dulu di-hardcode (LPG_3KG/PERTALITE):
// kini subsidi hanyalah salah satu Service, dan layanan lain (pendaftaran klinik,
// bantuan kelurahan, dll) cukup ditambahkan sebagai baris baru di tabel ini.
//
// Code adalah slug unik yang dikirim petugas saat klaim (mis. "LPG_3KG"). Inilah
// yang menjaga kompatibilitas: aplikasi mobile tetap mengirim "commodity" berisi Code.
type Service struct {
	ID   uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Code string    `gorm:"type:varchar(32);uniqueIndex;not null" json:"code"`
	Name string    `gorm:"type:varchar(128);not null" json:"name"`
	// Kind menentukan perilaku alur klaim: "quota" | "eligibility" | "log".
	//   - quota       : butuh kelayakan + potong kuota per periode (mis. subsidi).
	//   - eligibility : butuh kelayakan, hanya mencatat kejadian (tanpa kuota).
	//   - log         : sekadar verifikasi identitas + catat; semua warga terdaftar dilayani.
	Kind string `gorm:"type:varchar(16);not null" json:"kind"`
	// DefaultEligible: kelayakan default untuk layanan yang butuh kelayakan
	// (quota/eligibility) ketika warga BELUM punya baris ServiceEligibility.
	//   true  -> opt-out (semua warga layak kecuali dinonaktifkan) — perilaku subsidi lama.
	//   false -> opt-in  (tidak ada yang layak sampai didaftarkan) — untuk layanan tertarget.
	DefaultEligible bool `gorm:"not null;default:false" json:"default_eligible"`
	// IsActive: layanan nonaktif menolak klaim dengan HTTP 400 (kill-switch),
	// sama seperti dulu "komoditas tidak valid".
	IsActive  bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
