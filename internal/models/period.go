package models

import (
	"regexp"
	"time"
)

// WIB adalah zona waktu Indonesia Barat (UTC+7).
//
// Sengaja memakai FixedZone, BUKAN time.LoadLocation("Asia/Jakarta"), agar tidak
// bergantung pada ketersediaan database tzdata di runtime (mis. binary serverless
// Vercel yang ramping). Indonesia tidak menerapkan DST, jadi WIB selalu UTC+7
// dan offset tetap ini deterministik.
var WIB = time.FixedZone("WIB", 7*60*60)

// periodPattern memvalidasi format periode "YYYY-MM".
var periodPattern = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`)

// CurrentPeriod mengembalikan periode kuota berjalan dalam format "YYYY-MM"
// menurut waktu WIB.
//
// PENTING: periode HARUS dihitung dalam WIB, bukan UTC. Klaim yang terjadi antara
// pukul 00:00–07:00 WIB masih hari/bulan yang sama secara UTC dikurangi 7 jam —
// memakai UTC bisa membuat klaim mendekati pergantian bulan jatuh ke periode yang
// salah dan tertolak "kuota belum diatur" padahal kuotanya ada.
func CurrentPeriod() string {
	return time.Now().In(WIB).Format("2006-01")
}

// IsValidPeriod memastikan periode berformat "YYYY-MM" dengan bulan 01–12.
func IsValidPeriod(p string) bool {
	return periodPattern.MatchString(p)
}
