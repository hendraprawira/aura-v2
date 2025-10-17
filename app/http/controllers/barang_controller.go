package controllers

import (
	"github.com/goravel/framework/contracts/http"
)

type BarangController struct{}

// Menampilkan halaman login
func (a *BarangController) Index(ctx http.Context) http.Response {
	return ctx.Response().View().Make("data_barang/index.tmpl")
}
