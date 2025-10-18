package controllers

import (
	"aura/app/models"
	"fmt"
	"log"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct{}

func (a *AuthController) ShowLogin(ctx http.Context) http.Response {
	return ctx.Response().View().Make("login/login.tmpl")
}

func (a *AuthController) Login(ctx http.Context) http.Response {
	username := ctx.Request().Input("username")
	password := ctx.Request().Input("password")

	var user models.MasterUser

	if err := facades.Orm().Query().Where("username", username).First(&user); err != nil {
		fmt.Printf("Login gagal: User tidak ditemukan untuk username=%s. Error: %v\n", username, err)
		return ctx.Response().View().Make("login/login.tmpl", map[string]interface{}{
			"Error": "Username atau password salah",
		})
	}
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		fmt.Printf("Login gagal: Password salah untuk user=%s\n", user.Username)
		return ctx.Response().View().Make("login/login.tmpl", map[string]interface{}{
			"Error": "Username atau password salah",
		})
	}
	ctx.Response().Cookie(http.Cookie{Name: "user_id", Value: fmt.Sprintf("%d", user.ID), MaxAge: 3600})
	ctx.Response().Cookie(http.Cookie{Name: "username", Value: user.Username, MaxAge: 3600})
	ctx.Response().Cookie(http.Cookie{Name: "role", Value: user.Role, MaxAge: 3600})
	log.Printf("Login berhasil untuk user: %s\n", user.Username)
	return ctx.Response().Redirect(http.StatusFound, "/")
}
