package config

import (
	"os"
	"strconv"
	"time"
)

//À medida que o projeto cresce, ter várias funções GetXxx()
//  espalhadas fica difícil de gerenciar.
// Um padrão comum em Go é carregar tudo uma vez no startup

// Sugestões:
// nao existe limite máximo nas durações
// Um .env mal configurado com ACCESS_TOKEN_DURATION_MINUTES=99999 geraria tokens válidos
//
//	por ~69 dias sem nenhum aviso.
//
// Adicione um teto
func GetAccessTokenDuration() time.Duration {
	//Se alguém configurar ACCESS_TOKEN_DURATION_MINUTES=abc,
	// o sistema usa o fallback sem nenhum aviso. Em produção isso pode ser difícil de debugar
	val := os.Getenv("ACCESS_TOKEN_DURATION_MINUTES")
	//Os valores padrão aparecem duas vezes em cada função (no if val == "" e no if err).
	//  Uma constante elimina isso
	if val == "" {
		return 15 * time.Minute
	}
	minutes, err := strconv.Atoi(val)
	if err != nil || minutes <= 0 {
		return 15 * time.Minute
	}
	return time.Duration(minutes) * time.Minute
}

func GetRefreshTokenDuration() time.Duration {
	val := os.Getenv("REFRESH_TOKEN_DURATION_DAYS")
	if val == "" {
		return 7 * 24 * time.Hour
	}
	days, err := strconv.Atoi(val)
	if err != nil || days <= 0 {
		return 7 * 24 * time.Hour
	}
	return time.Duration(days) * 24 * time.Hour
}
