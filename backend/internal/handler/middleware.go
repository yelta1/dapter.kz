package handler

import (
	"context"
	"net/http"
	"strings"

	"dapter-kz/internal/config"
	"dapter-kz/internal/models"
	"dapter-kz/internal/service"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	userIDCtxKey contextKey = "userID"
	phoneCtxKey  contextKey = "phone"
	roleCtxKey   contextKey = "role"
)

// AuthMiddleware проверяет JWT токен и добавляет данные пользователя в контекст запроса
func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Отсутствует заголовок Authorization", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Неверный формат заголовка Authorization (ожидается: Bearer <token>)", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]
			claims := &service.Claims{}

			token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
				return []byte(cfg.JWTSecret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Недействительный или просроченный токен авторизации", http.StatusUnauthorized)
				return
			}

			// Добавляем claims в контекст
			ctx := context.WithValue(r.Context(), userIDCtxKey, claims.UserID)
			ctx = context.WithValue(ctx, phoneCtxKey, claims.Phone)
			ctx = context.WithValue(ctx, roleCtxKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole проверяет, соответствует ли роль пользователя в контексте одной из разрешенных
func RequireRole(allowedRoles ...models.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(roleCtxKey).(models.UserRole)
			if !ok {
				http.Error(w, "Роль пользователя не найдена в контексте", http.StatusForbidden)
				return
			}

			roleMatch := false
			for _, role := range allowedRoles {
				if userRole == role {
					roleMatch = true
					break
				}
			}

			if !roleMatch {
				http.Error(w, "Доступ запрещен для данной роли пользователя", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserIDFromContext вспомогательный метод для получения ID пользователя из контекста
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(userIDCtxKey).(string)
	return val, ok
}

// GetUserRoleFromContext вспомогательный метод для получения роли пользователя из контекста
func GetUserRoleFromContext(ctx context.Context) (models.UserRole, bool) {
	val, ok := ctx.Value(roleCtxKey).(models.UserRole)
	return val, ok
}

// GetUserPhoneFromContext вспомогательный метод для получения номера телефона из контекста
func GetUserPhoneFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(phoneCtxKey).(string)
	return val, ok
}

// CorsMiddleware добавляет CORS-заголовки и обрабатывает предзапросы OPTIONS
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Если это предварительный запрос (OPTIONS), завершаем его успешно
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
