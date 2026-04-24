package main

import (
	"os"
)

type Config struct {
	Port     string
	DBDsn    string
	EnableDB bool
}

func LoadConfig() Config {
	dbDsn := os.Getenv("DB_DSN")
	return Config{
		Port:     getEnv("PORT", "8081"),
		DBDsn:    dbDsn,
		EnableDB: dbDsn != "",
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
