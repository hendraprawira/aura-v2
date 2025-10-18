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
	authController := controllers.AuthController{}
	barangController := controllers.BarangController{}
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
	facades.Route().Get("/data-barang", RequireLogin(barangController.Index))

	facades.Route().Get("/logout", func(ctx http.Context) http.Response {
		ctx.Response().Cookie(http.Cookie{Name: "user_id", Value: "", MaxAge: -1})
		ctx.Response().Cookie(http.Cookie{Name: "username", Value: "", MaxAge: -1})
		ctx.Response().Cookie(http.Cookie{Name: "role", Value: "", MaxAge: -1})
		return ctx.Response().Redirect(http.StatusFound, "/login")
	})
	facades.Route().Get("/api/data-barang", barangController.DatatablesAPI)

}
