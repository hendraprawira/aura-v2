package controllers

import (
	"aura/app/models"
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

	query := facades.Orm().Query().Model(&models.JasaService{}).
		Where("is_deleted", false).
		WhereIn("kode_status", statusesAny).
		OrderBy("id", "desc")

	if err := query.Find(&services); err != nil {
		facades.Log().Error(err)
		services = []models.JasaService{}
	}

	return ctx.Response().View().Make("jasa_services/index.tmpl", map[string]any{
		"username":    username,
		"role":        role,
		"activeGroup": "pelayanan",
		"activeMenu":  "pelayanan",
		"menu":        "Jasa Services",
		"jam":         now,
		"ServiceList": services,
	})
}
