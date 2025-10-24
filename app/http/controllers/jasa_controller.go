package controllers

import (
	"aura/app/models"
	"strconv"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"gorm.io/gorm"
)

type JasaController struct{}

func (b *JasaController) Index(ctx http.Context) http.Response {
	username := ctx.Request().Cookie("username")
	role := ctx.Request().Cookie("role")
	if username == "" {
		return ctx.Response().Redirect(http.StatusFound, "/login")
	}
	return ctx.Response().View().Make("data_jasa/index.tmpl", map[string]any{
		"username":    username,
		"role":        role,
		"activeGroup": "master-data",
		"activeMenu":  "data-jasa",
		"menu":        "Data Jasa",
	})
}

// Data mengambil data jasa untuk DataTables (API endpoint)
func (b *JasaController) Data(ctx http.Context) http.Response {
	// Ambil parameter DataTables
	draw := ctx.Request().QueryInt("draw", 1)
	start := ctx.Request().QueryInt("start", 0)
	length := ctx.Request().QueryInt("length", 10)
	search := ctx.Request().Query("search[value]", "")
	orderColIndex := ctx.Request().QueryInt("order[0][column]", 1)
	orderDir := ctx.Request().Query("order[0][dir]", "asc")

	// Mapping kolom DataTables ke kolom database
	// Pastikan urutan dan nama ini sesuai dengan DataTables Anda
	columnMap := []string{"id", "kode_jasa", "nama_jasa", "harga_jasa", "harga_toko", "harga_mekanik", "keterangan", "id"}
	orderBy := columnMap[orderColIndex] + " " + orderDir

	var jasaList []models.DataJasa

	db := facades.Orm().Query().Model(&models.DataJasa{}).Where("is_deleted = ?", false)

	// 1. Hitung Total Records (tanpa filter/search)
	totalRecords, _ := facades.Orm().Query().Model(&models.DataJasa{}).Where("is_deleted", 0).Count()

	if search != "" {
		words := strings.Fields(search) // Pisah input berdasarkan spasi

		for _, word := range words {
			word = strings.TrimSpace(word)
			if word != "" {
				// Tambahkan kondisi AND untuk setiap kata
				db = db.Where(
					"kode_jasa LIKE ? OR nama_jasa LIKE ? OR keterangan LIKE ?",
					"%"+word+"%",
					"%"+word+"%",
					"%"+word+"%",
				)
			}
		}
	}

	// 3. Hitung Filtered Records
	filteredRecords, _ := db.Count()

	// 4. Ambil Data dengan Limit, Offset, dan Order
	// Lakukan Find() pada instance DB yang sudah difilter dan diberi limit/offset
	err := db.Limit(length).Offset(start).Order(orderBy).Find(&jasaList)
	if err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{"error": "Gagal mengambil data", "message": err.Error()})
	}
	// Format data untuk DataTables
	data := make(map[string]any)
	data["draw"] = draw
	data["recordsTotal"] = totalRecords
	data["recordsFiltered"] = filteredRecords
	data["data"] = jasaList

	return ctx.Response().Json(http.StatusOK, data)
}

// Store menyimpan data jasa baru
func (b *JasaController) Store(ctx http.Context) http.Response {
	userID, _ := strconv.Atoi(ctx.Request().Cookie("user_id"))

	var dto JasaDTO
	if err := ctx.Request().Bind(&dto); err != nil {
		return ctx.Response().Json(http.StatusUnprocessableEntity, http.Json{"message": "Validasi gagal", "errors": err.Error()})
	}

	jasa := models.DataJasa{
		KodeJasa:     dto.KodeJasa,
		NamaJasa:     dto.NamaJasa,
		HargaJasa:    dto.HargaJasa,
		HargaToko:    dto.HargaToko,
		HargaMekanik: dto.HargaMekanik,
		Keterangan:   dto.Keterangan,
		CreatedBy:    userID,
		UpdatedBy:    userID,
		IsDeleted:    false,
	}

	if err := facades.Orm().Query().Create(&jasa); err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{"message": "Gagal menyimpan data", "error": err.Error()})
	}

	return ctx.Response().Json(http.StatusCreated, http.Json{"message": "Data jasa berhasil ditambahkan", "data": jasa})
}

// Show menampilkan detail jasa untuk form edit
func (b *JasaController) Show(ctx http.Context) http.Response {
	idStr := ctx.Request().Route("jasa")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return ctx.Response().Json(http.StatusNotFound, http.Json{"message": "ID tidak valid"})
	}

	var jasa models.DataJasa
	err = facades.Orm().Query().Where("id", id).Where("is_deleted", 0).First(&jasa)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return ctx.Response().Json(http.StatusNotFound, http.Json{"message": "Data jasa tidak ditemukan"})
		}
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{"message": "Gagal mengambil data", "error": err.Error()})
	}

	return ctx.Response().Json(http.StatusOK, jasa)
}

// Update memperbarui data jasa yang sudah ada
func (b *JasaController) Update(ctx http.Context) http.Response {
	idStr := ctx.Request().Route("jasa")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return ctx.Response().Json(http.StatusUnprocessableEntity, http.Json{"message": "ID tidak valid"})
	}

	var dto JasaDTO
	if err := ctx.Request().Bind(&dto); err != nil {
		return ctx.Response().Json(http.StatusUnprocessableEntity, http.Json{"message": "Validasi gagal", "errors": err.Error()})
	}

	// ASUMSI: Ambil ID pengguna saat ini
	userID, _ := strconv.Atoi(ctx.Request().Cookie("user_id"))

	// Cek apakah data ada sebelum update
	var existingJasa models.DataJasa
	if err := facades.Orm().Query().Where("id", id).Where("is_deleted", 0).First(&existingJasa); err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return ctx.Response().Json(http.StatusNotFound, http.Json{"message": "Data jasa tidak ditemukan"})
		}
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{"message": "Gagal mencari data", "error": err.Error()})
	}

	updates := map[string]any{
		"kode_jasa":     dto.KodeJasa,
		"nama_jasa":     dto.NamaJasa,
		"harga_jasa":    dto.HargaJasa,
		"harga_toko":    dto.HargaToko,
		"harga_mekanik": dto.HargaMekanik,
		"keterangan":    dto.Keterangan,
		"updated_by":    userID,
	}

	// Update langsung
	if _, err := facades.Orm().Query().Model(&models.DataJasa{}).Where("id", id).Update(updates); err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{"message": "Gagal memperbarui data", "error": err.Error()})
	}

	return ctx.Response().Json(http.StatusOK, http.Json{"message": "Data jasa berhasil diperbarui"})
}

// Destroy menghapus (soft delete) data jasa
func (b *JasaController) Destroy(ctx http.Context) http.Response {
	idStr := ctx.Request().Route("jasa")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return ctx.Response().Json(http.StatusNotFound, http.Json{"message": "ID tidak valid"})
	}
	userID, _ := strconv.Atoi(ctx.Request().Cookie("user_id"))

	// Cek apakah data ada sebelum hapus
	var existingJasa models.DataJasa
	if err := facades.Orm().Query().Where("id", id).Where("is_deleted", 0).First(&existingJasa); err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return ctx.Response().Json(http.StatusNotFound, http.Json{"message": "Data jasa tidak ditemukan"})
		}
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{"message": "Gagal mencari data", "error": err.Error()})
	}

	// Lakukan Soft Delete
	updates := map[string]any{
		"is_deleted": 1,
		"deleted_by": userID,
		// Tambahkan DeletedAt jika Anda tidak menggunakan GORM's Gorm.DeletedAt field
		"deleted_at": time.Now(),
	}

	if _, err := facades.Orm().Query().Model(&models.DataJasa{}).Where("id", id).Update(updates); err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{"message": "Gagal menghapus data", "error": err.Error()})
	}

	return ctx.Response().Json(http.StatusOK, http.Json{"message": "Data jasa berhasil dihapus"})
}

// DataTransferObject untuk validasi input
type JasaDTO struct {
	// ID tidak perlu ada di DTO, karena hanya untuk Create/Update data
	KodeJasa     string `json:"kode_jasa" binding:"required"`
	NamaJasa     string `json:"nama_jasa" binding:"required"`
	HargaJasa    int    `json:"harga_jasa" binding:"required"`
	HargaToko    string `json:"harga_toko"`
	HargaMekanik string `json:"harga_mekanik"`
	Keterangan   string `json:"keterangan"`
}
