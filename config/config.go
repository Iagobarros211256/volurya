package config

import (
	"os"
	"strconv"
	"time"
)

func GetAccessTokenDuration() time.Duration {
	val := os.Getenv("ACCESS_TOKEN_DURATION_MINUTES")
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
