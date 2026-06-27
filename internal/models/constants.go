package models

// Role petugas/pengguna sistem.
const (
	RoleAdmin    = "admin"    // pegawai kelurahan/pemerintah (dashboard web)
	RoleMerchant = "merchant" // petugas lapangan (SPBU/pangkalan gas, dll)
)

// Kind layanan (Service.Kind) — menentukan perilaku alur klaim.
const (
	ServiceKindQuota       = "quota"       // butuh kelayakan + potong kuota per periode
	ServiceKindEligibility = "eligibility" // butuh kelayakan, hanya catat kejadian
	ServiceKindLog         = "log"         // sekadar verifikasi identitas + catat
)

// Code layanan bawaan yang di-seed. LPG_3KG & PERTALITE dipertahankan persis agar
// aplikasi mobile (yang mengirim "commodity" berisi nilai ini) tetap berfungsi.
const (
	ServiceCodeLPG3KG    = "LPG_3KG"
	ServiceCodePertalite = "PERTALITE"
	ServiceCodeKlinik    = "KLINIK_DAFTAR" // contoh layanan non-subsidi (Kind=log)
)

// Status hasil transaksi.
const (
	TxStatusSuccess  = "success"
	TxStatusRejected = "rejected"
)

// Alasan penolakan transaksi (dipakai di field Reason).
const (
	ReasonNotRegistered = "KTP tidak terdaftar"
	ReasonNotEligible   = "Warga tidak layak menerima layanan ini"
	ReasonQuotaEmpty    = "Kuota layanan sudah habis"
	ReasonNoQuotaConfig = "Kuota untuk layanan ini belum diatur"
)

// IsValidServiceKind memvalidasi Kind layanan.
func IsValidServiceKind(k string) bool {
	switch k {
	case ServiceKindQuota, ServiceKindEligibility, ServiceKindLog:
		return true
	default:
		return false
	}
}

// IsValidRole memvalidasi role pengguna dari request.
func IsValidRole(r string) bool {
	switch r {
	case RoleAdmin, RoleMerchant:
		return true
	default:
		return false
	}
}

// IsValidTxStatus memvalidasi status transaksi (dipakai untuk filter monitoring).
func IsValidTxStatus(s string) bool {
	switch s {
	case TxStatusSuccess, TxStatusRejected:
		return true
	default:
		return false
	}
}
