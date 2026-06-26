package models

// Role petugas/pengguna sistem.
const (
	RoleAdmin    = "admin"    // pegawai kelurahan/pemerintah (dashboard web)
	RoleMerchant = "merchant" // petugas lapangan (SPBU/pangkalan gas)
)

// Jenis komoditas subsidi yang bisa diklaim.
const (
	CommodityLPG3KG    = "LPG_3KG"
	CommodityPertalite = "PERTALITE"
)

// Status hasil transaksi klaim.
const (
	TxStatusSuccess  = "success"
	TxStatusRejected = "rejected"
)

// Alasan penolakan transaksi (dipakai di field Reason).
const (
	ReasonNotRegistered = "KTP tidak terdaftar"
	ReasonNotEligible   = "Warga tidak layak menerima subsidi"
	ReasonQuotaEmpty    = "Kuota subsidi sudah habis"
	ReasonNoQuotaConfig = "Kuota untuk komoditas ini belum diatur"
)

// IsValidCommodity memvalidasi jenis komoditas dari request.
func IsValidCommodity(c string) bool {
	switch c {
	case CommodityLPG3KG, CommodityPertalite:
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
