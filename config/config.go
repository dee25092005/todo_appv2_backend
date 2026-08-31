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
	R2          R2Config
}

type R2Config struct {
	AccountID       string
	AccessKeyID     string
	AccessKeySecret string
	BucketName      string
	PublicURL       string
	Region          string
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
		R2: R2Config{
			AccountID:       getEnv("R2_ACCOUNT_ID", ""),
			AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
			AccessKeySecret: getEnv("R2_ACCESS_KEY_SECRET", ""),
			BucketName:      getEnv("R2_BUCKET_NAME", ""),
			PublicURL:       getEnv("R2_PUBLIC_URL", ""),
			Region:          getEnv("R2_REGION", ""),
		},
	}

	if cfg.Port == "" {
		log.Fatal("PORT is not set")
	}

	if cfg.DatabaseUrl == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	if cfg.R2.AccessKeyID == "" {
		log.Fatal("R2_ACCESS_KEY_ID is not set")
	}

	if cfg.R2.AccessKeySecret == "" {
		log.Fatal("R2_ACCESS_KEY_SECRET is not set")
	}

	if cfg.R2.BucketName == "" {
		log.Fatal("R2_BUCKET_NAME is not set")
	}
	if cfg.R2.PublicURL == "" {
		log.Fatal("R2_PUBLIC_URL is not set")
	}
	if cfg.R2.Region == "" {
		log.Fatal("R2_REGION is not set")
	}

	return cfg, nil
}
