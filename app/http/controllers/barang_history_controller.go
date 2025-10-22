package controllers

import (
	"aura/app/models"
	"fmt" // Import fmt untuk formatting string

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

type BarangHistoryController struct{}

// Struct Response yang diformat untuk dikirim ke template
type HistoryResponse struct {
	No                 int // Nomor urut
	KodeBarang         string
	NamaBarang         string
	KodeTranksaksi     string
	SourceChangeBarang string
	QtyAwal            int
	// QtyTranksaksiTampil akan berisi string "+10" atau "-5"
	QtyTranksaksiTampil string
	QtyTampilClass      string // Field baru untuk kelas CSS
	QtyTerakhir         int
	TanggalPerubahan    string // Sudah dalam format string (WIB)
	CreatedName         string
}

// Index menampilkan halaman history, mengambil semua data history untuk di-looping di template
func (h *BarangHistoryController) Index(ctx http.Context) http.Response {
	username := ctx.Request().Cookie("username")
	role := ctx.Request().Cookie("role")
	barangID := ctx.Request().Route("id")

	// 1. Ambil Data Barang utama (untuk mendapatkan Kode Barang)
	var barang models.DataBarang
	err := facades.Orm().Query().Select("kode_item", "nama_item").Find(&barang, barangID)

	if err != nil {
		return ctx.Response().View().Make("error.404")
	}

	kodeBarang := barang.KodeItem

	// 2. Ambil semua Data History
	var historyList []models.DataBarangHistory
	facades.Orm().Query().
		Model(&models.DataBarangHistory{}).
		Where("kode_barang = ? AND is_deleted = ?", kodeBarang, false).
		Order("id desc"). // Urutan DESC seperti permintaan
		Find(&historyList)

	// 3. Mapping dan Formatting Data (Nomor urut, QTY Tranksaksi, dan Tanggal)
	var formattedHistoryList []HistoryResponse
	for i, h := range historyList {
		qtyTampil := fmt.Sprintf("%d", h.QtyTambahKurang)
		qtyClass := "text-danger"
		// Menambahkan tanda '+' jika IsTambah = true
		if h.IsTambah {
			qtyTampil = fmt.Sprintf("+%d", h.QtyTambahKurang)
			qtyClass = "text-success"
		}

		formatted := HistoryResponse{
			No:                  i + 1, // Nomor urut dimulai dari 1
			KodeBarang:          h.KodeBarang,
			NamaBarang:          h.NamaBarang,
			KodeTranksaksi:      h.KodeTranksaksi,
			SourceChangeBarang:  h.SourceChangeBarang,
			QtyAwal:             h.QtyAwal,
			QtyTranksaksiTampil: qtyTampil, // Sudah diformat
			QtyTerakhir:         h.QtyTerakhir,
			TanggalPerubahan:    formatWIB(h.TanggalPerubahan), // Sudah diformat
			QtyTampilClass:      qtyClass,                      // Set kelas
			CreatedName:         h.CreatedName,
		}
		formattedHistoryList = append(formattedHistoryList, formatted)
	}

	// 4. Persiapkan data yang akan dikirim ke template
	data := map[string]interface{}{
		"BarangID":    barangID,
		"KodeBarang":  barang.KodeItem,
		"NamaBarang":  barang.NamaItem,
		"History":     formattedHistoryList, // Kirim data yang sudah diformat
		"username":    username,
		"role":        role,
		"activeGroup": "master-data",
		"activeMenu":  "data-jasa",
		"menu":        "Data Jasa",
	}

	return ctx.Response().View().Make("data_barang/history.tmpl", data)
}
