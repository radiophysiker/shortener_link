package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/radiophysiker/shortener_link/internal/config"
)

type contextKey string

const UserIDContextKey contextKey = "userID"

// Claims структура для JWT токена
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// AuthMiddleware проверяет наличие и валидность JWT токена в куки
// Если токен отсутствует, создает новый и устанавливает его в куки
// Добавляет userID в контекст запроса
func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := checkJWTToken(r, cfg)
			if err != nil {
				zap.L().Error("failed to get userID from JWT", zap.Error(err))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if userID == "" {
				userID = uuid.New().String()
				setJWTCookie(w, userID, cfg)
			}

			// Добавляем userID в контекст
			ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// checkJWTToken извлекает и валидирует userID из JWT токена с кук
func checkJWTToken(r *http.Request, cfg *config.Config) (userID string, err error) {
	cookie, err := r.Cookie(cfg.CookieName)
	if err != nil {
		return "", nil
	}

	tokenString := cookie.Value
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecretKey), nil
	})

	if err != nil {
		zap.L().Warn("invalid JWT token", zap.Error(err))
		return "", nil
	}

	if token == nil {
		zap.L().Warn("nil JWT token")
		return "", nil
	}
	if !token.Valid {
		zap.L().Warn("invalid JWT token")
		return "", nil
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		zap.L().Warn("invalid JWT claims")
		return "", nil
	}

	if claims.UserID == "" {
		zap.L().Warn("userID is empty in JWT claims")
		return "", fmt.Errorf("userID is empty in JWT claims")
	}
	return claims.UserID, nil
}

// setJWTCookie создает и устанавливает JWT токен в куки
func setJWTCookie(w http.ResponseWriter, userID string, cfg *config.Config) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "shortener-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecretKey))
	if err != nil {
		zap.L().Error("failed to sign JWT token", zap.Error(err))
		return
	}

	cookie := &http.Cookie{
		Name:     cfg.CookieName,
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour),
	}

	http.SetCookie(w, cookie)
}

// GetUserIDFromContext извлекает userID из контекста запроса
func GetUserIDFromContext(ctx context.Context) string {
	userID, ok := ctx.Value(UserIDContextKey).(string)
	if !ok {
		return ""
	}
	return userID
}
