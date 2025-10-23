package models

import (
	"github.com/goravel/framework/database/orm"
)

// DataJasa merepresentasikan baris dalam tabel 'data_jasa'
type DataJasa struct {
	orm.Model

	// Tambahkan JSON Tag untuk mengontrol output JSON
	KodeJasa     string `gorm:"column:kode_jasa" json:"kode_jasa"`
	NamaJasa     string `gorm:"column:nama_jasa" json:"nama_jasa"`
	HargaJasa    int    `gorm:"column:harga_jasa" json:"harga_jasa"`
	HargaToko    string `gorm:"column:harga_toko" json:"harga_toko"`
	HargaMekanik string `gorm:"column:harga_mekanik" json:"harga_mekanik"`
	Keterangan   string `gorm:"column:keterangan" json:"keterangan"`
	CreatedBy    int    `gorm:"column:created_by" json:"created_by"`
	UpdatedBy    int    `gorm:"column:updated_by" json:"updated_by"`
	DeletedBy    int    `gorm:"column:deleted_by" json:"deleted_by"`
	IsDeleted    bool   `gorm:"column:is_deleted" json:"is_deleted"`
}

// TableName mengoverride nama tabel default Goravel/GORM
func (DataJasa) TableName() string {
	return "data_jasa"
}
