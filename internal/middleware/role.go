package middleware

import (
	"net/http"
	"ssubench/internal/domain"
	"ssubench/internal/handler"
)

func RoleChecker(predicate func(domain.UserRole) bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		checkRole := predicate

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := domain.UserRole(r.Context().Value(domain.UserRoleKey).(string))

			if !checkRole(userRole) {
				handler.SendError(w, http.StatusForbidden, "нет доступа")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
