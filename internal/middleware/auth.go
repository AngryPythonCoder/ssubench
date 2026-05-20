package middleware

import (
	"context"
	"net/http"
	"ssubench/internal/domain"
	"ssubench/internal/handler"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				handler.SendError(w, http.StatusUnauthorized, "отсутствует заголовок авторизации")
				return
			}

			parts := strings.Split(authHeader, " ")

			if len(parts) != 2 || parts[0] != "Bearer" {
				handler.SendError(w, http.StatusUnauthorized, "неверный формат токена")
				return
			}

			tokenString := parts[1]

			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				handler.SendError(w, http.StatusUnauthorized, "недопустимый токен")
				return
			}

			value, ok := claims["user_id"].(float64)
			if !ok {
				handler.SendError(w, http.StatusUnauthorized, "недопустимый токен")
				return
			}
			userID := int(value)

			userRole, ok := claims["role"].(string)
			if !ok {
				handler.SendError(w, http.StatusUnauthorized, "недопустимый токен")
				return
			}

			ctx := context.WithValue(r.Context(), domain.UserIDKey, userID)
			ctx = context.WithValue(ctx, domain.UserRoleKey, userRole)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
