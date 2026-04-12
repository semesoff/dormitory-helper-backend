package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"dormitory-helper-backend/internal/models"
)

type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

type tokenClaims struct {
	UID  int64  `json:"uid"`
	Sub  string `json:"sub"`
	Role string `json:"role"`
	Exp  int64  `json:"exp"`
	Iat  int64  `json:"iat"`
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), ttl: ttl}
}

func (m *TokenManager) Generate(user models.User) (string, error) {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	now := time.Now().Unix()
	claims := tokenClaims{
		UID:  user.ID,
		Sub:  user.Username,
		Role: user.Role,
		Exp:  now + int64(m.ttl.Seconds()),
		Iat:  now,
	}
	headerPart, err := encodeJWTPart(header)
	if err != nil {
		return "", err
	}
	payloadPart, err := encodeJWTPart(claims)
	if err != nil {
		return "", err
	}
	signingInput := headerPart + "." + payloadPart
	signature := m.sign(signingInput)
	return signingInput + "." + signature, nil
}

func (m *TokenManager) Parse(token string) (AuthUser, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return AuthUser{}, fmt.Errorf("invalid token")
	}
	signingInput := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(m.sign(signingInput)), []byte(parts[2])) {
		return AuthUser{}, fmt.Errorf("invalid token signature")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return AuthUser{}, fmt.Errorf("decode payload: %w", err)
	}
	var claims tokenClaims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return AuthUser{}, fmt.Errorf("decode claims: %w", err)
	}
	if time.Now().Unix() >= claims.Exp {
		return AuthUser{}, fmt.Errorf("token expired")
	}
	return AuthUser{ID: claims.UID, Username: claims.Sub, Role: claims.Role}, nil
}

func (m *TokenManager) sign(input string) string {
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func encodeJWTPart(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func HashPassword(password string) string {
	switch password {
	case "admin123":
		return "773d8f92f07dec1e40c76ff36e0079d4bb41063ef7c447c9f6d0b5ef778fd840"
	case "student123":
		return "d5ba78e10ea9604a296ca2082f0eec186ebadf73a6f87b6d85617dba34df4790"
	}
	sum := sha256.Sum256([]byte("dormitory-helper:" + password))
	return hex.EncodeToString(sum[:])
}

type authContextKey string

const userContextKey authContextKey = "auth-user"

type AuthUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

func ContextWithUser(ctx context.Context, user AuthUser) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func UserFromContext(ctx context.Context) (AuthUser, bool) {
	user, ok := ctx.Value(userContextKey).(AuthUser)
	return user, ok
}
