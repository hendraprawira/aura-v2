package models

import "time"

type DataBarang struct {
	ID           int        `gorm:"column:id;primaryKey;autoIncrement"`
	KodeItem     string     `gorm:"column:kode_item"`
	NamaItem     string     `gorm:"column:nama_item"`
	Merk         string     `gorm:"column:merk"`
	Stok         int        `gorm:"column:stok"`
	Satuan       string     `gorm:"column:satuan"`
	Rak          string     `gorm:"column:rak"`
	HargaPokok   int        `gorm:"column:harga_pokok"`
	HargaJual    int        `gorm:"column:harga_jual"`
	HargaToko    int        `gorm:"column:harga_toko"`
	HargaOrang   int        `gorm:"column:harga_orang"`
	HargaBengkel int        `gorm:"column:harga_bengkel"`
	StokMaksimal int        `gorm:"column:stok_maksimal"`
	StokMinimal  int        `gorm:"column:stok_minimal"`
	Keterangan   string     `gorm:"column:keterangan"`
	SkuBarang    string     `gorm:"column:sku_barang"`
	KodeBarcode  string     `gorm:"column:kode_barcode"`
	CreatedBy    int        `gorm:"column:created_by"`
	CreatedAt    *time.Time `gorm:"column:created_at"`
	UpdatedBy    int        `gorm:"column:updated_by"`
	UpdatedAt    *time.Time `gorm:"column:updated_at"`
	DeletedBy    int        `gorm:"column:deleted_by"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
	IsDeleted    bool       `gorm:"column:is_deleted"`

	RowNum int `gorm:"-"`
}

func (DataBarang) TableName() string {
	return "data_barang"
}

type DatatablesResponse struct {
	Draw            int         `json:"draw"`
	RecordsTotal    int64       `json:"recordsTotal"`
	RecordsFiltered int64       `json:"recordsFiltered"`
	Data            interface{} `json:"data"`
}
