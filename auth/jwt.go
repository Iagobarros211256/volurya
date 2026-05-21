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
	//secret como variável global mutável
	//Funciona,
	// mas o padrão mais seguro é encapsular num struct
	// para facilitar testes e rotação de chave futuramente:
	secret = []byte(secretStr)
}

// Access e Refresh token usam o mesmo secret e mesma estrutura de Claims
// Isso é um risco sério.
// Um refresh token poderia ser usado onde se espera um access token e vice-versa.
// lembrar de Adicionar um campo TokenType nas claims ou use um secret diferente para cada tipo
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
			//time.Now() chamado múltiplas vezes na mesma função
			//Detalhe menor, mas IssuedAt, NotBefore e ExpiresAt chamam time.Now() separadamente.
			// Capture uma vez
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			// Opcional: Issuer: "volurya-api",
			// Opcional: Audience: jwt.ClaimStrings{"api"},
			//sempre esqueco de ativar esses campos Os campos opcionais comentados valem a pena ativar.
			//Issuer e Audience adicionam uma camada extra de validação sem custo nenhum. Se no futuro
			// houver mais de um serviço consumindo tokens, isso evita que um token de um contexto seja aceito em outro.
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
	//slog.Error dentro de ValidateToken é verboso demais
	//Toda falha de token (token expirado de usuário,
	//token inválido) vai logar um erro.
	// Em produção com tráfego alto isso polui os logs.
	// Erros esperados deveriam ser slog.Debug ou slog.Warn,
	// reservando Error para falhas sistêmicas.
	//como nao sei se de fato minha api vai ter
	// esse trtafego todo no futuro decidi manter like this
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
