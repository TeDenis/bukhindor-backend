package app

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// GenerateUUID генерирует новый UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// HashPassword хеширует пароль с помощью bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash проверяет пароль против хеша
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// HashToken хеширует токен с помощью SHA256 (для токенов, не паролей)
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// GenerateRandomToken генерирует случайный токен заданной длины
func GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ValidateEmail проверяет корректность email
func ValidateEmail(email string) bool {
	if len(email) > MaxEmailLength {
		return false
	}
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// ValidatePassword проверяет корректность пароля
func ValidatePassword(password string) bool {
	return len(password) >= MinPasswordLength && len(password) <= MaxPasswordLength
}

// ValidateName проверяет корректность имени
func ValidateName(name string) bool {
	name = strings.TrimSpace(name)
	return len(name) > 0 && len(name) <= MaxNameLength
}

// ValidateAppType проверяет корректность типа приложения
func ValidateAppType(appType string) bool {
	switch appType {
	case AppTypeIOS, AppTypeAndroid, AppTypeWeb:
		return true
	default:
		return false
	}
}

// IsExpired проверяет, истекло ли время
func IsExpired(expiresAt time.Time) bool {
	return time.Now().After(expiresAt)
}
