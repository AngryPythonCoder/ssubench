package middleware

import (
	"fmt"
	"net/http"
	"ssubench/internal/domain"
	"ssubench/internal/handler"
)

func RoleChecker(role domain.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		errorMessage := fmt.Sprintf("пользователь должен иметь роль %v", string(role))

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := domain.UserRole(r.Context().Value(domain.UserRoleKey).(string))

			if userRole != role {
				handler.SendError(w, http.StatusUnauthorized, errorMessage)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
