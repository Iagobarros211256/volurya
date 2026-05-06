package auth

import (
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secret []byte

func init() {
	secretStr := os.Getenv("JWT_SECRET")
	if secretStr == "" {
		slog.Error("JWT_SECRET environment variable is required")
		os.Exit(1)
	}
	secret = []byte(secretStr)
}

type Claims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// create a token

func GenerateToken(userID int, role string, duration time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			// Opcional: Issuer: "volurya-api",
			// Opcional: Audience: jwt.ClaimStrings{"api"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})

	if err != nil {
		slog.Error("JWT parse failed", "error", err)
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		slog.Debug("JWT valid", "user_id", claims.UserID, "role", claims.Role)
		return claims, nil
	}

	slog.Warn("JWT invalid claims or expired")
	return nil, errors.New("token invalid or expired")
}

func GenerateRefreshToken(userID int, role string, duration time.Duration) (string, time.Time, error) {
	expiresAt := time.Now().Add(duration)

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}
