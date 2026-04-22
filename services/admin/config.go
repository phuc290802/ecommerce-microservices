package main

import (
	"os"
	"time"
)

type Config struct {
	Port           string
	DBDsn          string
	JWTSecret      string
	AccessTokenTTL time.Duration
}

func LoadConfig() Config {
	return Config{
		Port:           getEnv("PORT", "8088"),
		DBDsn:          getEnv("DB_DSN", "root:root@tcp(mysql:3306)/ecommerce?parseTime=true"),
		JWTSecret:      getEnv("JWT_SECRET", "supersecret"),
		AccessTokenTTL: 24 * time.Hour,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
