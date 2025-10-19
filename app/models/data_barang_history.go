package models

import (
	"time"
)

type DataBarangHistory struct {
	ID                 int        `gorm:"primaryKey;column:id"`
	KodeBarang         string     `gorm:"column:kode_barang"`
	NamaBarang         string     `gorm:"column:nama_barang"`
	SourceChangeBarang string     `gorm:"column:source_change_barang"`
	QtyTambahKurang    int        `gorm:"column:qty_tambah_kurang"`
	QtyAwal            int        `gorm:"column:qty_awal"`
	QtyTerakhir        int        `gorm:"column:qty_terakhir"`
	KodeTranksaksi     string     `gorm:"column:kode_tranksaksi"`
	IsTambah           bool       `gorm:"column:is_tambah"`
	IsKurang           bool       `gorm:"column:is_kurang"`
	IsTranksaksi       bool       `gorm:"column:is_tranksaksi"`
	TanggalPerubahan   *time.Time `gorm:"column:tanggal_perubahan"`
	CreatedName        string     `gorm:"column:created_name"`
	CreatedBy          int        `gorm:"column:created_by"`
	CreatedAt          *time.Time `gorm:"column:created_at"`
	UpdatedBy          int        `gorm:"column:updated_by"`
	UpdatedAt          *time.Time `gorm:"column:updated_at"`
	DeletedBy          int        `gorm:"column:deleted_by"`
	DeletedAt          *time.Time `gorm:"column:deleted_at"`
	IsDeleted          bool       `gorm:"column:is_deleted"`
}

// TableName overrides the table name
func (DataBarangHistory) TableName() string {
	return "data_barang_history"
}
