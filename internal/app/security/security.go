package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Константы, используемые в пакете.
const (
	// TokenExp — время жизни JWT-токена (по умолчанию 3 часа).
	TokenExp = time.Hour * 3

	// AuthorizationTokenName — имя куки, в которой хранится токен аутентификации.
	AuthorizationTokenName = "Authorization"
)

// Глобальные ошибки, используемые в пакете.
var (
	// ErrNoAuthorizationToken — ошибка, возникающая, когда токен отсутствует.
	ErrNoAuthorizationToken = errors.New("authorization token is missed")
)

// Claims — пользовательские claims для JWT-токена.
// Содержит стандартные поля и уникальный идентификатор пользователя.
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

// GetToken извлекает значение токена из куки запроса.
//
// Параметры:
//   - req: указатель на http.Request.
//
// Возвращает:
//   - string: значение токена или пустую строку, если токен не найден.
func GetToken(req *http.Request) string {
	tokenCookie, err := req.Cookie(AuthorizationTokenName)
	if tokenCookie != nil && err == nil {
		return tokenCookie.Value
	}
	return ""
}

// GetUserID извлекает идентификатор пользователя из JWT-токена.
//
// Параметры:
//   - token: строковое представление JWT-токена.
//   - encodedKey: ключ для верификации токена.
//
// Возвращает:
//   - string: идентификатор пользователя.
//   - error: nil, если успешно, иначе — ошибку.
func GetUserID(token string, encodedKey string) (string, error) {
	if token == "" {
		return "", nil
	}
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(encodedKey), nil
	})
	if err != nil {
		return "", fmt.Errorf("get user id: claims parsing: %w", err)
	}
	if claims.UserID == "" {
		return "", ErrNoAuthorizationToken
	}
	return claims.UserID, nil
}

// BuildToken создаёт новый JWT-токен с уникальным идентификатором пользователя.
//
// Параметры:
//   - key: секретный ключ для подписи токена.
//
// Возвращает:
//   - string: готовый токен.
//   - error: nil, если успешно, иначе — ошибку.
func BuildToken(key string) (string, error) {
	userID, err := generateUserID()
	if err != nil {
		return "", fmt.Errorf("generate new user id: %w", err)
	}
	tokenWithClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: userID,
	})
	token, err := tokenWithClaims.SignedString([]byte(key))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return token, nil
}

// CreateHashForURL создаёт уникальный хэш для заданного URL.
//
// Параметры:
//   - fullURL: оригинальный URL.
//
// Возвращает:
//   - string: хэшированное представление URL (первые 10 символов).
//   - error: nil, если URL корректный, иначе — ошибку.
func CreateHashForURL(fullURL string) (string, error) {
	if isInvalidURL(fullURL) {
		return "", errors.New("invalid url")
	}
	return CreateHash(fullURL), nil
}

// CreateHash создаёт SHA256-хэш строки и возвращает его URL-safe представление.
//
// Параметры:
//   - key: строка, которую нужно захэшировать.
//
// Возвращает:
//   - string: первые 10 символов Base64-URL-encoded хэша.
func CreateHash(key string) string {
	hashBytes := sha256.Sum256([]byte(key))
	hashString := base64.URLEncoding.EncodeToString(hashBytes[:])
	return hashString[:10]
}

// IsAllowedTrustedIP проверяет, входит ли IP-адрес клиента из заголовка X-Real-IP
// в доверенную подсеть, заданную в trustedSubnet.
//
// Параметры:
//   - req: HTTP-запрос для получения заголовка X-Real-IP.
//   - res: HTTP-ответ, куда записываются ошибки с кодом 403 Forbidden при проверках.
//   - trustedSubnet: строка CIDR с доверенной подсетью (например, "192.168.1.0/24").
//
// Возвращает:
//   - bool: true, если IP корректен и входит в подсеть;
//     false и пишет ошибку в ответ при отсутствии или некорректности данных.
func IsAllowedTrustedIP(req *http.Request, res http.ResponseWriter, trustedSubnet string) bool {
	if trustedSubnet == "" {
		http.Error(res, "TrustedSubnet is missing", http.StatusForbidden)
		return false
	}
	ipStr := req.Header.Get("X-Real-IP")
	if ipStr == "" {
		http.Error(res, "X-Real-IP header missing", http.StatusForbidden)
		return false
	}
	ip := net.ParseIP(strings.TrimSpace(ipStr))
	if ip == nil {
		http.Error(res, "Invalid X-Real-IP", http.StatusForbidden)
		return false
	}
	_, subnet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		http.Error(res, "Invalid TRUSTED_SUBNET", http.StatusForbidden)
		return false
	}
	if !subnet.Contains(ip) {
		http.Error(res, "Forbidden", http.StatusForbidden)
		return false
	}
	return true
}

// encrypt шифрует строку с использованием AES-GCM.
//
// Параметры:
//   - token: строка, которую нужно зашифровать.
//   - key: hex-строка, используемая как ключ шифрования.
//
// Возвращает:
//   - string: зашифрованный токен в hex-представлении.
//   - error: nil, если успешно, иначе — ошибку.
func encrypt(token string, key string) (string, error) {
	decodedKey, err := hex.DecodeString(key)
	if err != nil {
		return "", fmt.Errorf("decode key: %w", err)
	}

	keyHash := sha256.Sum256(decodedKey)
	aesblock, err := aes.NewCipher(keyHash[:])
	if err != nil {
		return "", fmt.Errorf("new aes block: %w", err)
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return "", fmt.Errorf("new aes gcm: %w", err)
	}

	nonce := keyHash[len(keyHash)-aesgcm.NonceSize():]
	encryptedToken := aesgcm.Seal(nil, nonce, []byte(token), nil)
	return hex.EncodeToString(encryptedToken), nil
}

// generateUserID генерирует случайную строку длиной 16 байт, используемую как идентификатор пользователя.
//
// Возвращает:
//   - string: случайная строка.
//   - error: nil, если успешно, иначе — ошибку.
func generateUserID() (string, error) {
	userID, err := generateRandom(16)
	if err != nil {
		return "", fmt.Errorf("generate random: %w", err)
	}
	return string(userID), nil
}

// generateRandom генерирует случайную последовательность байтов заданной длины.
//
// Параметры:
//   - size: количество байтов.
//
// Возвращает:
//   - []byte: случайная последовательность.
//   - error: nil, если успешно, иначе — ошибку.
func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// isInvalidURL проверяет, является ли URL допустимым.
//
// Проверяет наличие scheme и host.
//
// Параметры:
//   - fullURL: URL для проверки.
//
// Возвращает:
//   - bool: true, если URL невалиден, иначе — false.
func isInvalidURL(fullURL string) bool {
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return true
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return true
	}
	return false
}
