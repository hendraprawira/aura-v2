package middlewares

import (
	"github.com/goravel/framework/contracts/http"
)

type AuthMiddleware struct{}

func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

func (m *AuthMiddleware) Handle(ctx http.Context) {
	// 1. Ambil Cookie yang menunjukkan status login
	userID := ctx.Request().Cookie("user_id")

	// 2. Cek apakah cookie ada atau ada error saat mengambilnya
	if userID == "" {
		// Jika tidak ada cookie atau cookie kosong, user belum login.
		// Arahkan ke halaman login

		// Opsional: Simpan URL yang diminta agar setelah login bisa kembali ke halaman tersebut
		// facades.Session().Put(ctx.Request().Context(), "intended_url", ctx.Request().Url())

		ctx.Response().Redirect(http.StatusFound, "/login") // Ganti /login sesuai route Anda
		return                                              // Hentikan eksekusi handler controller
	}

	// 3. (Opsional tapi direkomendasikan): Validasi ID user ke database
	// Anda bisa menambahkan logika di sini untuk memastikan user_id valid dan user masih aktif.
	// Jika tidak valid, lakukan redirect ke /login.

	// Lanjutkan ke Controller Handler jika sudah login
	ctx.Request().Next()
}
