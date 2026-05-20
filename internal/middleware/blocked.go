package middleware

import (
	"net/http"
	"ssubench/internal/domain"
	"ssubench/internal/handler"
	"ssubench/internal/service"
)

func BlockChecker(service *service.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(domain.UserIDKey).(int)

			user, err := service.Get(r.Context(), userID)
			if err != nil {
				handler.SendError(w, http.StatusUnauthorized, "пользователь не найден")
				return
			}

			if user.Status == domain.StatusBlocked {
				handler.SendError(w, http.StatusUnauthorized, "пользователь заблокирован")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
