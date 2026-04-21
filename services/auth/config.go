package main

import (
	"os"
	"time"
)

type Config struct {
	Port            string
	DBDsn           string
	RedisAddr       string
	JWTSecret       string
	RefreshTokenTTL time.Duration
	ResetTokenTTL   time.Duration
	OTPTokenTTL     time.Duration
}

func LoadConfig() Config {
	return Config{
		Port:            getEnv("PORT", "8084"),
		DBDsn:           getEnv("DB_DSN", "user:password@tcp(mysql:3306)/ecommerce?charset=utf8mb4&parseTime=true"),
		RedisAddr:       getEnv("REDIS_ADDR", "redis:6379"),
		JWTSecret:       getEnv("JWT_SECRET", "supersecret"),
		RefreshTokenTTL: 7 * 24 * time.Hour,
		ResetTokenTTL:   15 * time.Minute,
		OTPTokenTTL:     5 * time.Minute,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
