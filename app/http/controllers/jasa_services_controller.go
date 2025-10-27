package controllers

import (
	"aura/app/models"
	"log"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

type JasaServicesController struct{}

func (h *JasaServicesController) Index(ctx http.Context) http.Response {
	username := ctx.Request().Cookie("username")
	role := ctx.Request().Cookie("role")
	now := time.Now().Format("02-01-2006")

	var services []models.JasaService

	// ✅ Convert []int ke []any untuk WhereIn
	statuses := []int{0, 1}
	statusesAny := make([]any, len(statuses))
	for i, v := range statuses {
		statusesAny[i] = v
	}

	// ✅ Query utama untuk data jasa service
	query := facades.Orm().Query().Model(&models.JasaService{}).
		Where("is_deleted", false).
		WhereIn("kode_status", statusesAny).
		OrderBy("id", "desc")

	totalServiceProcess, _ := facades.Orm().Query().Model(&models.JasaService{}).
		Where("is_deleted = ?", false).Where("kode_status = ?", 0).Count()

	totalServiceDone, _ := facades.Orm().Query().Model(&models.JasaService{}).
		Where("is_deleted = ?", false).Where("kode_status = ?", 1).Count()

	if err := query.Find(&services); err != nil {
		facades.Log().Error(err)
		services = []models.JasaService{}
	}

	// ✅ Mekanik Join Query
	type MekanikData struct {
		NamaMekanik  string `gorm:"column:nama_mekanik"`
		KodeServices string `gorm:"column:kode_services"`
	}

	var mekanikData []MekanikData

	if err := facades.Orm().Query().
		Table("data_services").
		Join("services_mekanik", "data_services.kode_services", "=", "services_mekanik.kode_services").
		Where("data_services.is_deleted = ?", false).
		WhereIn("data_services.kode_status", statusesAny).
		Where("services_mekanik.is_deleted = ?", false).
		Select("services_mekanik.nama_mekanik, data_services.kode_services").
		OrderBy("data_services.id", "desc").
		Get(&mekanikData); err != nil {
		facades.Log().Error("Error get mekanik data:", err)
	}

	// ✅ Grouping dan konsolidasi nama mekanik
	groupedMekanik := map[string][]string{}
	for _, m := range mekanikData {
		groupedMekanik[m.KodeServices] = append(groupedMekanik[m.KodeServices], m.NamaMekanik)
	}

	finalMekanik := map[string]map[string]any{}
	for kode, names := range groupedMekanik {
		uniqueNames := uniqueStrings(names)
		if len(uniqueNames) > 2 {
			finalMekanik[kode] = map[string]any{
				"kode_services": kode,
				"nama_mekanik":  joinStrings(uniqueNames, ", "),
			}
		} else {
			finalMekanik[kode] = map[string]any{
				"kode_services": kode,
				"nama_mekanik":  uniqueNames,
			}
		}
	}
	log.Print(finalMekanik)

	// ✅ Kirim semua data ke template
	return ctx.Response().View().Make("jasa_services/index.tmpl", map[string]any{
		"username":            username,
		"role":                role,
		"activeGroup":         "pelayanan",
		"activeMenu":          "pelayanan",
		"menu":                "Jasa Services",
		"jam":                 now,
		"ServiceList":         services,
		"totalServiceProcess": totalServiceProcess,
		"totalServiceDone":    totalServiceDone,
		"finalMekanik":        finalMekanik,
	})
}

// ✅ Helper: hilangkan duplikat string
func uniqueStrings(list []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, s := range list {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// ✅ Helper: join string slice dengan separator
func joinStrings(list []string, sep string) string {
	result := ""
	for i, s := range list {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
