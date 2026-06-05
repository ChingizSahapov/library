package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

// Config хранит параметры запуска сервиса.
type Config struct {
	ServerPort string
	DBPath     string
}

// Load читает .env (если есть) и переменные окружения.
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Info("no .env file found, using environment variables")
	}

	cfg := &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		DBPath:     getEnv("DB_PATH", "./library.db"),
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
