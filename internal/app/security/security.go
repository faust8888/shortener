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

func GetToken(req *http.Request) string {
	tokenCookie, err := req.Cookie(AuthorizationTokenName)
	if tokenCookie != nil && err == nil {
		return tokenCookie.Value
	}
	return ""
}

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
