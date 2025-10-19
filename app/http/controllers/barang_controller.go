package controllers

import (
	"aura/app/models"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

type BarangController struct{}

func (b *BarangController) Index(ctx http.Context) http.Response {
	return ctx.Response().View().Make("data_barang/index.tmpl")
}

func (a *BarangController) DatatablesAPI(ctx http.Context) http.Response {
	request := ctx.Request()
	draw, _ := strconv.Atoi(request.Query("draw"))
	start, _ := strconv.Atoi(request.Query("start"))   // offset
	length, _ := strconv.Atoi(request.Query("length")) // limit/records per page
	searchVal := request.Query("search[value]")

	// 2. Tentukan Kolom Sorting
	orderByColIndex, _ := strconv.Atoi(request.Query("order[0][column]"))
	orderDir := request.Query("order[0][dir]") // asc/desc

	// Sesuaikan nama kolom DB berdasarkan indeks kolom Datatables
	// HATI-HATI: Urutan ini harus sama persis dengan urutan columns di JS Datatables
	columnNames := []string{
		"",              // 0. No (tidak di-sort)
		"kode_item",     // 1. Kode Barang
		"nama_item",     // 2. Nama Barang
		"merk",          // 3. Merk
		"stok",          // 4. Stok
		"harga_pokok",   // 5. Harga Pokok
		"harga_jual",    // 6. Harga Jual Ecer
		"stok_minimal",  // 7. Stok Min.
		"stok_maksimal", // 8. Stok Max.
		"keterangan",    // 9. Ket
		"updated_at",    // 10. Latest Update
		"",              // 11. Aksi (tidak di-sort)
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
	// Clone query untuk menghitung jumlah setelah filter, lalu apply count
	recordsFiltered, _ := db.Count()

	// 6. Ambil Data dengan Pagination dan Sorting
	var barangList []models.DataBarang
	if orderByColumn != "" { // Pastikan kolom bisa di-sort
		db = db.Order(orderClause)
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
		// Jika tidak ditemukan atau error database
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
	user_id, _ := strconv.Atoi(ctx.Request().Cookie("user_id"))

	now := time.Now()
	barangID := ctx.Request().Route("id")
	if barangID == "" {
		return ctx.Response().Json(http.StatusBadRequest, map[string]string{"message": "ID Barang tidak ditemukan"})
	}

	var req UpdateBarangRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, map[string]string{"message": "Gagal membaca JSON: " + err.Error()})
	}

	body, _ := io.ReadAll(ctx.Request().Origin().Body)
	fmt.Println("RAW BODY:", string(body))
	ctx.Request().Bind(&req)
	fmt.Printf("PARSED REQUEST: %+v\n", req)
	fmt.Print(ctx.Request())

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
		UpdatedBy:    user_id,
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
			CreatedBy:          user_id,  // ID user yang login
		}

		facades.Orm().Query().Create(&history) // Simpan history (Error diabaikan untuk penyederhanaan)
	}

	return ctx.Response().Json(http.StatusOK, map[string]string{"message": "Data barang berhasil diubah"})
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
