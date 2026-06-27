package models

// SeedServices adalah daftar layanan bawaan platform — SATU sumber kebenaran yang
// dipakai BAIK oleh cmd/migrate (backfill) MAUPUN cmd/seed, agar definisinya tidak
// pernah divergen (mis. salah ketik Code di dua tempat).
//
// LPG_3KG & PERTALITE memakai DefaultEligible=true untuk meniru perilaku subsidi lama
// (warga baru otomatis layak). Layanan baru sebaiknya DefaultEligible=false (opt-in).
func SeedServices() []Service {
	return []Service{
		{Code: ServiceCodeLPG3KG, Name: "LPG 3 Kg", Kind: ServiceKindQuota, DefaultEligible: true, IsActive: true},
		{Code: ServiceCodePertalite, Name: "Pertalite", Kind: ServiceKindQuota, DefaultEligible: true, IsActive: true},
		{Code: ServiceCodeKlinik, Name: "Pendaftaran Klinik", Kind: ServiceKindLog, DefaultEligible: false, IsActive: true},
	}
}
