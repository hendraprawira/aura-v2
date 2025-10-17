package routes

import (
	"aura/app/http/controllers"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/route"
)

func Web() {
	// ✅ Serve static files from ./public/assets
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
			"username": username,
			"role":     role,
		})
	})
	// facades.Route().Get("/data-barang", barangController.Index)
	facades.Route().Group(func(router route.Route) {
		router.Middleware(facades.Route().Middleware("auth")).
			Prefix("/barang").
			Get("/", new(controllers.BarangController).Index)
1``	})
}
