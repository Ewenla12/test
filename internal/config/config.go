package config

import "os"

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
}

func Load() Config {
	return Config{
		DatabaseURL: getEnv("DATABASE_URL"),
		JWTSecret:   getEnv("JWT_SECRET"),
		Port:        getEnv("PORT"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
