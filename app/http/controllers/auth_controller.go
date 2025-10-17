package controllers

import (
	"aura/app/models"
	"fmt"
	"log"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"golang.org/x/crypto/bcrypt" // Pastikan library ini di-import
)

type AuthController struct{}

// Menampilkan halaman login
func (a *AuthController) ShowLogin(ctx http.Context) http.Response {
	return ctx.Response().View().Make("login/login.tmpl")
}

// Proses login yang menggunakan bcrypt untuk keamanan
func (a *AuthController) Login(ctx http.Context) http.Response {
	username := ctx.Request().Input("username")
	password := ctx.Request().Input("password")

	var user models.MasterUser

	// 1. Ambil data user dari database berdasarkan username.
	// Jika user tidak ditemukan, First() akan mengembalikan error dan kita langsung
	// mengarahkan ke halaman login dengan pesan error generik.
	if err := facades.Orm().Query().Where("username", username).First(&user); err != nil {
		// Penting: Jangan tampilkan detail error database ke user, cukup log saja.
		fmt.Printf("Login gagal: User tidak ditemukan untuk username=%s. Error: %v\n", username, err)
		return ctx.Response().View().Make("login/login.tmpl", map[string]interface{}{
			"Error": "Username atau password salah",
		})
	}

	// 2. Lakukan pengecekan password yang aman menggunakan bcrypt.
	// Diasumsikan password di database (user.Password) sudah dalam bentuk hash bcrypt.
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		// Jika password tidak cocok, CompareHashAndPassword akan mengembalikan error.
		fmt.Printf("Login gagal: Password salah untuk user=%s\n", user.Username)
		return ctx.Response().View().Make("login/login.tmpl", map[string]interface{}{
			"Error": "Username atau password salah",
		})
	}

	// 3. Login berhasil → langsung redirect tanpa session
	// simpan cookie user
	ctx.Response().Cookie(http.Cookie{Name: "user_id", Value: fmt.Sprintf("%d", user.ID), MaxAge: 3600})
	ctx.Response().Cookie(http.Cookie{Name: "username", Value: user.Username, MaxAge: 3600})
	ctx.Response().Cookie(http.Cookie{Name: "role", Value: user.Role, MaxAge: 3600})
	log.Printf("Login berhasil untuk user: %s\n", user.Username)
	// PERBAIKAN: Mengganti StatusAccepted (202) menjadi StatusFound (302) untuk redirect yang benar.
	return ctx.Response().Redirect(http.StatusFound, "/")
}
