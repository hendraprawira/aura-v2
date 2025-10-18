package controllers

import (
	"aura/app/models"
	"strconv"
	"strings"

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

	// 7. Bentuk Respons JSON Datatables
	response := models.DatatablesResponse{
		Draw:            draw,
		RecordsTotal:    totalRecords,
		RecordsFiltered: recordsFiltered,
		Data:            barangList,
	}

	return ctx.Response().Json(http.StatusOK, response)
}

// AJAX search / load more
func (b *BarangController) Search(ctx http.Context) http.Response {
	query := ctx.Request().Input("query", "")
	page := ctx.Request().InputInt("page", 1)
	perPage := ctx.Request().InputInt("per_page", 10)

	dbQuery := facades.Orm().Query().Table("data_barang").Where("is_deleted", 0)
	if query != "" {
		dbQuery = dbQuery.Where("nama_item", "LIKE", "%"+query+"%").
			OrWhere("kode_item", "LIKE", "%"+query+"%").
			OrWhere("kode_barcode", "LIKE", "%"+query+"%").
			OrWhere("merk", "LIKE", "%"+query+"%").
			OrWhere("sku_barang", "LIKE", "%"+query+"%")
	}

	total, _ := dbQuery.Count()

	var barangList []models.DataBarang
	dbQuery.OrderBy("id", "desc").
		Offset((page - 1) * perPage).
		Limit(perPage).
		Get(&barangList)

	return ctx.Response().Json(200, map[string]any{
		"barangList": barangList,
		"total":      total,
		"page":       page,
		"perPage":    perPage,
	})
}
