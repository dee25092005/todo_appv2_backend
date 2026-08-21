package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseUrl string
	JWTSecret   string
	JWTExp      int
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return value
	}
	return fallback
}

func LoadConfig() (*Config, error) {

	_ = godotenv.Load()
	cfg := &Config{
		Port:        getEnv("PORT", ""),
		DatabaseUrl: getEnv("DATABASE_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", ""),
		JWTExp:      getEnvAsInt("JWT_EXP", 0),
	}

	if cfg.Port == "" {
		log.Fatal("PORT is not set")
	}

	if cfg.DatabaseUrl == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	return cfg, nil
}
