package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	userIDCookieName = "user_id"
	secretKey        = "super-secret-key" // В продакшене должен быть в конфигурации
)

type userIDKeyType string

const userIDKey userIDKeyType = "userID"

// GetUserIDFromContext извлекает UserID из контекста
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}

// AuthMiddleware добавляет аутентификацию через подписанные cookies
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := getUserIDFromCookie(r)

		// Если userID не найден или недействительна подпись, создаем новый
		if userID == "" {
			userID = generateUserID()
			setUserIDCookie(w, userID)
		}

		// Добавляем userID в контекст
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getUserIDFromCookie извлекает и проверяет UserID из cookie
func getUserIDFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(userIDCookieName)
	if err != nil {
		return ""
	}

	// Проверяем подпись
	if !verifySignature(cookie.Value) {
		zap.L().Debug("Invalid cookie signature")
		return ""
	}

	// Извлекаем UserID из подписанного значения
	return extractUserIDFromSignedValue(cookie.Value)
}

// setUserIDCookie устанавливает подписанную cookie с UserID
func setUserIDCookie(w http.ResponseWriter, userID string) {
	signedValue := signValue(userID)
	cookie := &http.Cookie{
		Name:     userIDCookieName,
		Value:    signedValue,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((24 * time.Hour).Seconds()), // 24 часа
	}
	http.SetCookie(w, cookie)
}

// generateUserID генерирует новый уникальный идентификатор пользователя
func generateUserID() string {
	return uuid.New().String()
}

// signValue подписывает значение с помощью HMAC
func signValue(value string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(value))
	signature := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("%s.%s", value, signature)
}

// verifySignature проверяет подпись значения
func verifySignature(signedValue string) bool {
	userID := extractUserIDFromSignedValue(signedValue)
	if userID == "" {
		return false
	}
	expectedSignedValue := signValue(userID)
	return hmac.Equal([]byte(signedValue), []byte(expectedSignedValue))
}

// extractUserIDFromSignedValue извлекает UserID из подписанного значения
func extractUserIDFromSignedValue(signedValue string) string {
	// Находим последнюю точку (разделитель между значением и подписью)
	lastDotIndex := -1
	for i := len(signedValue) - 1; i >= 0; i-- {
		if signedValue[i] == '.' {
			lastDotIndex = i
			break
		}
	}

	if lastDotIndex == -1 || lastDotIndex == 0 || lastDotIndex == len(signedValue)-1 {
		return ""
	}

	return signedValue[:lastDotIndex]
}
