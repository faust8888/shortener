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
	"net/http"
	"net/url"
	"time"
)

const (
	TokenExp               = time.Hour * 3
	AuthorizationTokenName = "Authorization"
)

var ErrNoAuthorizationToken = errors.New("authorization token is missed")

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func GetToken(cookies []*http.Cookie) string {
	for _, cookie := range cookies {
		if cookie.Name == AuthorizationTokenName {
			return cookie.Value
		}
	}
	return ""
}

func GetUserID(encryptedToken string, encodedKey string) (string, error) {
	if encryptedToken == "" {
		return "", nil
	}
	token, err := decrypt(encryptedToken, encodedKey)
	if err != nil {
		return "", fmt.Errorf("get user id: %w", err)
	}
	claims := &Claims{}
	_, err = jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
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
	encryptedToken, err := encrypt(token, key)
	if err != nil {
		return "", fmt.Errorf("encrypt: %w", err)
	}
	return encryptedToken, nil
}

func CreateHashForURL(fullURL string) (string, error) {
	if isInvalidURL(fullURL) {
		return "", errors.New("invalid url")
	}
	return CreateHash(fullURL), nil
}

func CreateHash(key string) string {
	hashBytes := sha256.Sum256([]byte(key))
	hashString := base64.URLEncoding.EncodeToString(hashBytes[:])
	return hashString[:10]
}

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

func decrypt(encryptedToken string, encodedKey string) (string, error) {
	decodedKey, err := hex.DecodeString(encodedKey)
	if err != nil {
		return "", fmt.Errorf("decrypt: decode keyHash - %w", err)
	}

	keyHash := sha256.Sum256(decodedKey)
	aesblock, err := aes.NewCipher(keyHash[:])
	if err != nil {
		return "", fmt.Errorf("decrypt: new cipher - %w", err)
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return "", fmt.Errorf("decrypt: new gcm - %w", err)
	}

	nonce := keyHash[len(keyHash)-aesgcm.NonceSize():]
	decodedToken, err := hex.DecodeString(encryptedToken)
	if err != nil {
		return "", fmt.Errorf("decrypt: decode encryptedToken - %w", err)
	}

	decryptedToken, err := aesgcm.Open(nil, nonce, decodedToken, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(decryptedToken), nil
}

func generateUserID() (string, error) {
	userID, err := generateRandom(16)
	if err != nil {
		return "", fmt.Errorf("generate random: %w", err)
	}
	return string(userID), nil
}

func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

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
