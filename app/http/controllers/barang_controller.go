package controllers

import (
	"aura/app/models"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

type BarangController struct{}

// Views
func (b *BarangController) Index(ctx http.Context) http.Response {

	username := ctx.Request().Cookie("username")
	role := ctx.Request().Cookie("role")

	if username == "" {
		return ctx.Response().Redirect(http.StatusFound, "/login")
	}

	return ctx.Response().View().Make("data_barang/index.tmpl", map[string]any{
		"username":    username,
		"role":        role,
		"activeGroup": "master-data",
		"activeMenu":  "data-barang",
		"menu":        "Data Barang",
	})
}

func (a *BarangController) DatatablesAPI(ctx http.Context) http.Response {
	request := ctx.Request()
	draw, _ := strconv.Atoi(request.Query("draw"))
	start, _ := strconv.Atoi(request.Query("start"))
	length, _ := strconv.Atoi(request.Query("length"))
	searchVal := request.Query("search[value]")

	orderByColIndex, _ := strconv.Atoi(request.Query("order[0][column]"))
	orderDir := request.Query("order[0][dir]") // asc/desc

	// Sesuaikan nama kolom DB berdasarkan indeks kolom Datatables
	// HATI-HATI: Urutan ini harus sama persis dengan urutan columns di JS Datatables
	columnNames := []string{
		"",
		"kode_item",
		"nama_item",
		"merk",
		"stok",
		"harga_pokok",
		"harga_jual",
		"stok_minimal",
		"stok_maksimal",
		"keterangan",
		"updated_at",
		"", // 11. Aksi (tidak di-sort)
	}

	orderByColumn := columnNames[orderByColIndex]
	orderClause := orderByColumn + " " + orderDir

	// 3. Hitung Total Data (Tanpa Filter)
	totalRecords, _ := facades.Orm().Query().Model(&models.DataBarang{}).Where("is_deleted = ?", false).Count()

	// 4. Bangun Query dengan Filtering (Pencarian)
	db := facades.Orm().Query().Model(&models.DataBarang{}).Where("is_deleted = ?", false)

	if searchVal != "" {
		words := strings.Fields(searchVal) // Pisah input berdasarkan spasi

		for _, word := range words {
			word = strings.TrimSpace(word)
			if word != "" {
				// Tambahkan kondisi AND untuk setiap kata
				db = db.Where(
					"nama_item LIKE ? OR kode_item LIKE ? OR merk LIKE ? OR kode_barcode LIKE ? OR sku_barang LIKE ?",
					"%"+word+"%",
					"%"+word+"%",
					"%"+word+"%",
					"%"+word+"%",
					"%"+word+"%",
				)
			}
		}
	}

	// 5. Hitung Total Data Setelah Filter
	recordsFiltered, _ := db.Count()

	// 6. Ambil Data dengan Pagination dan Sorting
	var barangList []models.DataBarang

	if orderByColumn != "" { // Pastikan kolom bisa di-sort
		db = db.Order(orderClause)
	} else {
		db = db.Order("created_at DESC")
	}

	// Apply limit dan offset
	db.Offset(start).Limit(length).Find(&barangList)

	var formattedList []BarangResponse
	for _, b := range barangList {
		rowClass := ""
		if b.Stok == 0 {
			rowClass = "stock-empty" // Putih (Default, tapi bisa di-highlight jika perlu)
		} else if b.Stok < b.StokMinimal {
			rowClass = "stock-warning" // Kuning (Kurang dari minimal)
		} else {
			rowClass = "stock-safe" // Hijau (Aman, lebih besar dari minimal)
		}

		formatted := BarangResponse{

			ID:           b.ID,
			KodeItem:     b.KodeItem,
			NamaItem:     b.NamaItem,
			Merk:         b.Merk,
			Stok:         b.Stok,
			HargaPokok:   b.HargaPokok,
			HargaJual:    b.HargaJual,
			StokMinimal:  b.StokMinimal,
			StokMaksimal: b.StokMaksimal,
			Keterangan:   b.Keterangan,
			UpdatedAt:    formatWIB(b.UpdatedAt),
			RowClass:     rowClass,
		}
		formattedList = append(formattedList, formatted)
	}

	// 7. Bentuk Respons JSON Datatables
	response := models.DatatablesResponse{
		Draw:            draw,
		RecordsTotal:    totalRecords,
		RecordsFiltered: recordsFiltered,
		Data:            formattedList,
	}

	return ctx.Response().Json(http.StatusOK, response)
}

func (b *BarangController) DetailAPI(ctx http.Context) http.Response {
	// Ambil ID dari parameter URL, contoh: /api/data-barang/123/detail
	barangID := ctx.Request().Route("id")
	if barangID == "" {
		return ctx.Response().Json(http.StatusBadRequest, map[string]string{"message": "ID Barang tidak ditemukan"})
	}

	var barang models.DataBarang
	// Cari barang berdasarkan ID dan hanya ambil kolom harga yang diperlukan
	err := facades.Orm().Query().Select("harga_toko", "harga_orang", "harga_bengkel", "kode_item").Where("id = ? AND is_deleted = ?", barangID, false).First(&barang)

	if err != nil {
		return ctx.Response().Json(http.StatusNotFound, map[string]string{"message": "Data Barang tidak ditemukan"})
	}
	// Buat respons yang hanya berisi data harga spesifik
	response := map[string]interface{}{
		"kode_item":     barang.KodeItem,
		"harga_toko":    barang.HargaToko,
		"harga_orang":   barang.HargaOrang,
		"harga_bengkel": barang.HargaBengkel,
	}

	return ctx.Response().Json(http.StatusOK, response)
}

// Save To DB
func (b *BarangController) Store(ctx http.Context) http.Response {
	username := ctx.Request().Cookie("username")
	userID, _ := strconv.Atoi(ctx.Request().Cookie("user_id"))
	now := time.Now()

	var req CreateBarangRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, map[string]string{"message": "Gagal membaca JSON: " + err.Error()})
	}

	// 1. Validasi Kunci
	if req.KodeItem == "" || req.NamaItem == "" || req.HargaJual <= 0 {
		return ctx.Response().Json(http.StatusBadRequest, map[string]string{"message": "Kode Barang, Nama Barang, dan Harga Jual wajib diisi."})
	}

	// 2. Cek Duplikasi Kode Barang
	var existingBarang models.DataBarang
	facades.Orm().Query().Where("kode_item = ? AND is_deleted = ?", req.KodeItem, false).First(&existingBarang)
	if existingBarang.ID > 0 {
		return ctx.Response().Json(http.StatusConflict, map[string]string{"message": "Kode Barang sudah ada. Gunakan kode lain."})
	}

	// 3. Buat objek DataBarang baru
	newBarang := models.DataBarang{
		KodeItem:     req.KodeItem,
		NamaItem:     req.NamaItem,
		Merk:         req.Merk,
		Stok:         req.Stok,
		Satuan:       req.Satuan,
		Rak:          req.Rak,
		HargaPokok:   req.HargaPokok,
		HargaJual:    req.HargaJual,
		HargaToko:    req.HargaToko,
		HargaOrang:   req.HargaOrang,
		HargaBengkel: req.HargaBengkel,
		StokMaksimal: req.StokMaksimal,
		StokMinimal:  req.StokMinimal,
		Keterangan:   req.Keterangan,
		SkuBarang:    req.SkuBarang,
		KodeBarcode:  req.KodeBarcode,
		CreatedBy:    userID, // Gunakan user_id dari cookie
		CreatedAt:    &now,
		UpdatedBy:    userID,
		UpdatedAt:    &now,
		IsDeleted:    false,
	}

	// 4. Simpan ke Database
	if err := facades.Orm().Query().Create(&newBarang); err != nil {
		facades.Log().Error(fmt.Sprintf("Gagal menyimpan barang baru: %s", err.Error()))
		return ctx.Response().Json(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan data barang baru."})
	}

	// 5. Tambahkan History Stok Awal (Jika Stok > 0)
	if newBarang.Stok > 0 {
		history := models.DataBarangHistory{ // Asumsi model DataBarangHistory sudah ada
			KodeBarang:         newBarang.KodeItem,
			NamaBarang:         newBarang.NamaItem,
			SourceChangeBarang: "CREATE_FORM",
			QtyTambahKurang:    newBarang.Stok,
			QtyAwal:            0,
			QtyTerakhir:        newBarang.Stok,
			KodeTranksaksi:     "NEW-" + newBarang.KodeItem + "-" + time.Now().Format("060102150405"),
			IsTambah:           true,
			IsTranksaksi:       false,
			TanggalPerubahan:   &now,
			CreatedName:        username,
			CreatedBy:          userID,
		}
		facades.Orm().Query().Create(&history)
	}

	return ctx.Response().Json(http.StatusCreated, map[string]string{"message": "Data barang baru berhasil ditambahkan"})
}

// EditAPI: Mengambil data barang lengkap untuk ditampilkan di form edit
func (b *BarangController) EditAPI(ctx http.Context) http.Response {
	barangID := ctx.Request().Route("id")
	if barangID == "" {
		return ctx.Response().Json(http.StatusBadRequest, map[string]string{"message": "ID Barang tidak ditemukan"})
	}

	var barang models.DataBarang
	// Ambil SEMUA data yang diperlukan untuk form edit
	err := facades.Orm().Query().Where("id = ? AND is_deleted = ?", barangID, false).First(&barang)

	if err != nil {
		return ctx.Response().Json(http.StatusNotFound, map[string]string{"message": "Data Barang tidak ditemukan"})
	}

	// Kirim objek model DataBarang utuh (atau mapping jika perlu, tapi DataBarang sudah cukup)
	return ctx.Response().Json(http.StatusOK, barang)
}

// Update: Menyimpan perubahan data barang
func (b *BarangController) Update(ctx http.Context) http.Response {
	username := ctx.Request().Cookie("username")
	userID, _ := strconv.Atoi(ctx.Request().Cookie("user_id"))

	now := time.Now()
	barangID := ctx.Request().Route("id")
	if barangID == "" {
		return ctx.Response().Json(http.StatusBadRequest, map[string]string{"message": "ID Barang tidak ditemukan"})
	}

	var req UpdateBarangRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, map[string]string{"message": "Gagal membaca JSON: " + err.Error()})
	}

	// 1. Validasi Sederhana
	if req.NamaItem == "" || req.Stok < 0 || req.HargaPokok < 0 || req.HargaJual <= 0 {
		return ctx.Response().Json(http.StatusBadRequest, map[string]string{"message": "Data Nama Item, Stok, Harga Pokok, dan Harga Jual tidak boleh kosong/negatif."})
	}

	// 2. Ambil data barang saat ini (diperlukan untuk history, jika stok berubah)
	var oldBarang models.DataBarang
	err := facades.Orm().Query().Where("id = ?", barangID).First(&oldBarang)
	if err != nil {
		return ctx.Response().Json(http.StatusNotFound, map[string]string{"message": "Data Barang tidak ditemukan saat update."})
	}

	// 3. Update data barang (dan cek jika ada perubahan Stok untuk History)
	newStok := req.Stok
	stokChange := newStok - oldBarang.Stok

	// Data yang akan diupdate
	updateData := models.DataBarang{
		NamaItem:     req.NamaItem,
		Merk:         req.Merk,
		Stok:         newStok,
		HargaPokok:   req.HargaPokok,
		HargaJual:    req.HargaJual,
		HargaToko:    req.HargaToko,
		HargaOrang:   req.HargaOrang,
		HargaBengkel: req.HargaBengkel,
		StokMinimal:  req.StokMinimal,
		StokMaksimal: req.StokMaksimal,
		Keterangan:   req.Keterangan,
		UpdatedBy:    userID,
		UpdatedAt:    &now,
	}

	_, errUpdate := facades.Orm().Query().Model(&models.DataBarang{}).Where("id = ?", barangID).Update(updateData)

	if errUpdate != nil {
		return ctx.Response().Json(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan perubahan: " + errUpdate.Error()})
	}

	// 4. Tambahkan History Stok jika ada perubahan stok
	if stokChange != 0 {
		isTambah := stokChange > 0
		qtyAbs := stokChange
		if !isTambah {
			qtyAbs = -stokChange // Ambil nilai absolut positif
		}

		history := models.DataBarangHistory{
			KodeBarang:         oldBarang.KodeItem,
			NamaBarang:         req.NamaItem, // Nama yang baru
			SourceChangeBarang: "EDIT_FORM",  // Sumber perubahan (misalnya: EDIT_FORM)
			QtyTambahKurang:    qtyAbs,
			QtyAwal:            oldBarang.Stok,
			QtyTerakhir:        newStok,
			KodeTranksaksi:     "EDIT-" + oldBarang.KodeItem + "-" + time.Now().Format("060102150405"),
			IsTambah:           isTambah,
			IsKurang:           !isTambah,
			IsTranksaksi:       false, // Perubahan manual, bukan transaksi
			TanggalPerubahan:   &now,
			CreatedName:        username, // Ganti dengan nama user yang login
			CreatedBy:          userID,   // ID user yang login
		}

		facades.Orm().Query().Create(&history) // Simpan history (Error diabaikan untuk penyederhanaan)
	}

	return ctx.Response().Json(http.StatusOK, map[string]string{"message": "Data barang berhasil diubah"})
}

func (b *BarangController) Delete(ctx http.Context) http.Response {
	username := ctx.Request().Cookie("username")
	userID, _ := strconv.Atoi(ctx.Request().Cookie("user_id"))
	now := time.Now()

	barangID := ctx.Request().Route("id")
	if barangID == "" {
		return ctx.Response().Json(http.StatusBadRequest, map[string]string{"message": "ID Barang tidak ditemukan"})
	}

	// Data untuk soft delete
	deleteData := map[string]interface{}{
		"is_deleted": true,
		"updated_by": userID,
		"updated_at": &now,
		"deleted_at": &now,   // Tambahkan kolom deleted_at
		"deleted_by": userID, // Tambahkan kolom deleted_by
	}

	// Lakukan update (soft delete)
	res, err := facades.Orm().Query().Model(&models.DataBarang{}).Where("id = ?", barangID).Update(deleteData)

	if err != nil {
		facades.Log().Error(fmt.Sprintf("Gagal Soft Delete Barang ID %s: %s", barangID, err.Error()))
		return ctx.Response().Json(http.StatusInternalServerError, map[string]string{"message": "Gagal menghapus data: " + err.Error()})
	}

	if res.RowsAffected == 0 {
		return ctx.Response().Json(http.StatusNotFound, map[string]string{"message": "Data Barang tidak ditemukan atau sudah terhapus."})
	}

	// 5. Tambahkan History (Opsional, tapi disarankan untuk audit)
	// Ambil data barang yang baru dihapus untuk history
	var oldBarang models.DataBarang
	facades.Orm().Query().Where("id = ?", barangID).Find(&oldBarang) // Ambil data terakhir, mungkin perlu pengecekan error yang lebih baik

	if oldBarang.KodeItem != "" {
		history := models.DataBarangHistory{
			KodeBarang:         oldBarang.KodeItem,
			NamaBarang:         oldBarang.NamaItem,
			SourceChangeBarang: "DELETE_FORM",
			QtyTambahKurang:    oldBarang.Stok, // Catat stok yang 'hilang'
			QtyAwal:            oldBarang.Stok,
			QtyTerakhir:        0, // Stok jadi 0 setelah dihapus (logis)
			KodeTranksaksi:     "DEL-" + oldBarang.KodeItem + "-" + time.Now().Format("060102150405"),
			IsKurang:           true,
			IsTranksaksi:       false,
			TanggalPerubahan:   &now,
			CreatedName:        username,
			CreatedBy:          userID,
		}
		facades.Orm().Query().Create(&history)
	}

	return ctx.Response().Json(http.StatusOK, map[string]string{"message": "Data barang berhasil dihapus (soft delete)."})
}

// SearchSuggestAPI: Menghasilkan 5 saran barang berdasarkan query
func (b *BarangController) SearchSuggestAPI(ctx http.Context) http.Response {
	query := ctx.Request().Query("q")
	// Menggunakan TrimSpace untuk membersihkan query
	query = strings.TrimSpace(query)

	// 1. Validasi Panjang Query (Minimal 3 karakter efektif)
	if len(query) < 3 {
		return ctx.Response().Json(http.StatusOK, map[string]interface{}{"data": []models.DataBarang{}})
	}

	// 2. Inisialisasi Query Database
	db := facades.Orm().Query().Model(&models.DataBarang{}).Where("is_deleted = ?", false)

	// 3. Membangun WHERE Clause berdasarkan setiap kata dalam query
	words := strings.Fields(query) // Pisah input berdasarkan spasi

	for _, word := range words {
		word = strings.TrimSpace(word)
		if word != "" {
			// Tambahkan kondisi AND untuk setiap kata
			// Setiap kata harus dicari di kolom nama_item ATAU kode_item ATAU merk ATAU kode_barcode ATAU sku_barang
			db = db.Where(
				"(nama_item LIKE ? OR kode_item LIKE ? OR merk LIKE ? )",
				"%"+word+"%",
				"%"+word+"%",
				"%"+word+"%",
			)
		}
	}

	// 4. Eksekusi Query
	var barangList []models.DataBarang

	// Batasi 5 hasil dan hanya ambil kolom yang diperlukan
	err := db.Limit(5).
		Select("id, kode_item, nama_item, stok").
		Find(&barangList)

	if err != nil {
		facades.Log().Error(fmt.Sprintf("Error saat mencari saran barang: %s", err.Error()))
		// Kirim respons error yang sesuai
		return ctx.Response().Json(http.StatusInternalServerError, map[string]string{"message": "Gagal mencari data barang"})
	}

	return ctx.Response().Json(http.StatusOK, map[string]interface{}{"data": barangList})
}

// AddStock: Menambah stok barang dan mencatat history
func (b *BarangController) AddStock(ctx http.Context) http.Response {
	username := ctx.Request().Cookie("username")
	userID, _ := strconv.Atoi(ctx.Request().Cookie("user_id"))

	var req AddStockRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, map[string]string{"message": "Gagal membaca JSON: " + err.Error()})
	}

	// 1. Validasi
	if req.BarangID <= 0 || req.QtyTambah <= 0 {
		return ctx.Response().Json(http.StatusBadRequest, map[string]string{"message": "ID Barang dan Jumlah Stok Tambahan tidak valid."})
	}

	now := time.Now()

	// 2. Ambil data barang saat ini
	var oldBarang models.DataBarang
	err := facades.Orm().Query().Where("id = ? AND is_deleted = ?", req.BarangID, false).First(&oldBarang)
	if err != nil {
		return ctx.Response().Json(http.StatusNotFound, map[string]string{"message": "Data Barang tidak ditemukan."})
	}

	// 3. Hitung Stok Baru
	newStok := oldBarang.Stok + req.QtyTambah

	// 4. Update Stok di database
	updateData := map[string]interface{}{
		"stok":       newStok,
		"updated_by": userID,
		"updated_at": &now,
	}

	_, errUpdate := facades.Orm().Query().Model(&models.DataBarang{}).Where("id = ?", req.BarangID).Update(updateData)

	if errUpdate != nil {
		facades.Log().Error(fmt.Sprintf("Gagal Update Stok Barang ID %d: %s", req.BarangID, errUpdate.Error()))
		return ctx.Response().Json(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan penambahan stok: " + errUpdate.Error()})
	}

	// 5. Tambahkan History Stok
	history := models.DataBarangHistory{
		KodeBarang:         oldBarang.KodeItem,
		NamaBarang:         oldBarang.NamaItem,
		SourceChangeBarang: "ADD_STOCK_FORM",
		QtyTambahKurang:    req.QtyTambah,
		QtyAwal:            oldBarang.Stok,
		QtyTerakhir:        newStok,
		KodeTranksaksi:     "ADD-" + oldBarang.KodeItem + "-" + time.Now().Format("060102150405"),
		IsTambah:           true,
		IsTranksaksi:       false, // Perubahan manual
		TanggalPerubahan:   &now,
		CreatedName:        username,
		CreatedBy:          userID,
	}

	facades.Orm().Query().Create(&history)

	return ctx.Response().Json(http.StatusOK, map[string]string{"message": "Stok barang berhasil ditambahkan."})
}

// Struct untuk request penambahan stok
type AddStockRequest struct {
	BarangID  int `json:"barang_id"`
	QtyTambah int `json:"qty_tambah"`
	UpdatedBy int `json:"updated_by"` // Dari user yang login
}
type BarangResponse struct {
	ID           int    `json:"ID"`
	KodeItem     string `json:"KodeItem"`
	NamaItem     string `json:"NamaItem"`
	Merk         string `json:"Merk"`
	Stok         int    `json:"Stok"`
	HargaPokok   int    `json:"HargaPokok"`
	HargaJual    int    `json:"HargaJual"`
	StokMinimal  int    `json:"StokMinimal"`
	StokMaksimal int    `json:"StokMaksimal"`
	Keterangan   string `json:"Keterangan"`
	UpdatedAt    string `json:"UpdatedAt"`
	RowClass     string `json:"RowClass"`
}

type UpdateBarangRequest struct {
	NamaItem     string `json:"nama_item"`
	Merk         string `json:"merk"`
	Stok         int    `json:"stok"`
	HargaPokok   int    `json:"harga_pokok"`
	HargaJual    int    `json:"harga_jual"`
	HargaToko    int    `json:"harga_toko"`
	HargaOrang   int    `json:"harga_orang"`
	HargaBengkel int    `json:"harga_bengkel"`
	StokMinimal  int    `json:"stok_minimal"`
	StokMaksimal int    `json:"stok_maksimal"`
	Keterangan   string `json:"keterangan"`
	UpdatedBy    int    `json:"updated_by"`
}

type CreateBarangRequest struct {
	KodeItem     string `json:"kode_item"`
	NamaItem     string `json:"nama_item"`
	Merk         string `json:"merk"`
	Stok         int    `json:"stok"`
	Satuan       string `json:"satuan"`
	Rak          string `json:"rak"`
	HargaPokok   int    `json:"harga_pokok"`
	HargaJual    int    `json:"harga_jual"`
	HargaToko    int    `json:"harga_toko"`
	HargaOrang   int    `json:"harga_orang"`
	HargaBengkel int    `json:"harga_bengkel"`
	StokMaksimal int    `json:"stok_maksimal"`
	StokMinimal  int    `json:"stok_minimal"`
	Keterangan   string `json:"keterangan"`
	SkuBarang    string `json:"sku_barang"`
	KodeBarcode  string `json:"kode_barcode"`
	CreatedBy    int    `json:"created_by"`
}
