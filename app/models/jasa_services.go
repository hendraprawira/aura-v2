// app/models/service.go
package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type JasaService struct {
	orm.Model
	KodeServices     string    `gorm:"column:kode_services"`
	NoPolisi         string    `gorm:"column:no_polisi"`
	TypeMotor        string    `gorm:"column:type_motor"`
	TanggalServices  time.Time `gorm:"column:tanggal_services"`
	UangMuka         string    `gorm:"column:uang_muka"`
	MetodePembayaran string    `gorm:"column:metode_pembayaran"`
	BayarCash        string    `gorm:"column:bayar_cash"`
	BayarNonCash     string    `gorm:"column:bayar_non_cash"`
	KodeStatus       int       `gorm:"column:kode_status"`
	StatusServices   string    `gorm:"column:status_services"`
	TotalHarga       string    `gorm:"column:total_harga"`
	CreatedBy        int       `gorm:"column:created_by"` // Ini adalah ID Mekanik
	DiskonPersen     string    `gorm:"column:diskon_persen"`
	IsDiskon         bool      `gorm:"column:is_diskon"`
	DiskonNominal    string    `gorm:"column:diskon_nominal"`
	TotalDiskon      string    `gorm:"column:total_diskon"`
	Pramuniaga       string    `gorm:"column:pramuniaga"`
	Catatan          string    `gorm:"column:catatan"`
	NoHp             string    `gorm:"column:no_hp"`
}

func (JasaService) TableName() string {
	return "data_services"
}
