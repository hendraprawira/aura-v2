package routes

import (
	"aura/app/http/controllers"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

func RequireLogin(handler func(ctx http.Context) http.Response) func(ctx http.Context) http.Response {
	return func(ctx http.Context) http.Response {
		username := ctx.Request().Cookie("username")
		if username == "" {
			return ctx.Response().Redirect(http.StatusFound, "/login")
		}
		return handler(ctx)
	}
}

func Web() {
	facades.Route().Static("/assets", "./public/assets")

	api := facades.Route().Prefix("api")

	authController := controllers.AuthController{}
	barangController := controllers.BarangController{}
	barangHistoryController := controllers.BarangHistoryController{}

	facades.Route().Get("/login", authController.ShowLogin)
	facades.Route().Post("/login", authController.Login)

	facades.Route().Get("/", func(ctx http.Context) http.Response {

		username := ctx.Request().Cookie("username")
		role := ctx.Request().Cookie("role")

		if username == "" {
			return ctx.Response().Redirect(http.StatusFound, "/login")
		}
		return ctx.Response().View().Make("index.tmpl", map[string]any{
			"username":    username,
			"role":        role,
			"activeGroup": "dashboard",
			"activeMenu":  "dashboard",
			"menu":        "Dashboard",
		})
	})

	facades.Route().Get("/logout", func(ctx http.Context) http.Response {
		ctx.Response().Cookie(http.Cookie{Name: "user_id", Value: "", MaxAge: -1})
		ctx.Response().Cookie(http.Cookie{Name: "username", Value: "", MaxAge: -1})
		ctx.Response().Cookie(http.Cookie{Name: "role", Value: "", MaxAge: -1})
		return ctx.Response().Redirect(http.StatusFound, "/login")
	})

	facades.Route().Get("/data-barang", RequireLogin(barangController.Index))
	api.Put("/data-barang/{id}", barangController.Update)
	api.Get("/data-barang/{id}/edit", barangController.EditAPI)
	api.Get("/data-barang/{id}/detail", barangController.DetailAPI)
	api.Get("/data-barang", barangController.DatatablesAPI)

	facades.Route().Get("/data-barang/history/{id}", barangHistoryController.Index)

}
