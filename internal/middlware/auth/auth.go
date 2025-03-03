package auth

import (
	"net/http"

	service "github.com/ruslantos/go-musthave-diploma-tpl/internal/mart"
)

func AuthMiddleware(userService *service.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// при регистрации не проверяем
			if r.Method == http.MethodPost && r.URL.Path == "/api/user/register" {
				next.ServeHTTP(w, r)
				return
			}

			login, password, ok := r.BasicAuth()
			if !ok {
				http.Error(w, "Authorization required", http.StatusUnauthorized)
				return
			}

			ctx := r.Context()
			if !userService.Authenticate(ctx, login, password) {
				http.Error(w, "Invalid credentials", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
