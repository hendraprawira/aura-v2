package controllers

import (

	// Import fmt untuk formatting string

	"aura/app/models"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

type JasaServicesController struct{}

// Index menampilkan halaman history, mengambil semua data history untuk di-looping di template
func (h *JasaServicesController) Index(ctx http.Context) http.Response {
	// username := ctx.Request().Cookie("username")
	// role := ctx.Request().Cookie("role")
	// 4. Persiapkan data yang akan dikirim ke template

	var services []models.JasaService

	// Ambil data services yang sedang diproses (misalnya kode_status = 1)
	// dan preload (JOIN) data mekanik
	err := facades.Orm().Query().
		Where("kode_status", 0).
		Where("is_deleted", false).
		Find(&services)

	// data := map[string]interface{}{
	// 	"username":    username,
	// 	"role":        role,
	// 	"activeGroup": "pelayanan",
	// 	"activeMenu":  "jasa-services",
	// 	"menu":        "Jasa Services",
	// 	"ServiceList": services,
	// }

	if err != nil {
		facades.Log().Error(err)
		return ctx.Response().View().Make("jasa_services/index.tmpl", map[string]any{
			"menu":        "Jasa Services",
			"ServiceList": []models.JasaService{}, // Kirim array kosong jika error
		})
	}
	return ctx.Response().View().Make("jasa_services/index.tmpl", map[string]any{
		"menu":        "Jasa Services",
		"ServiceList": services,
	})
	// return ctx.Response().View().Make("jasa_services/index.tmpl", data)
}
