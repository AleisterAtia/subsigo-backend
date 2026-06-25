package models

import "gorm.io/gorm"

// Migrate menjalankan AutoMigrate untuk seluruh model.
// Aman dipanggil berulang kali (idempotent): membuat tabel/kolom/index yang belum ada,
// tanpa menghapus data yang sudah ada.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Citizen{},
		&SubsidyQuota{},
		&Transaction{},
	)
}
