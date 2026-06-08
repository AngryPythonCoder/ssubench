package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBURL              string
	JWTSecret          string
	JWTTTL             time.Duration
	PasswordCost       int
	StartingBalance    int
	MaxPaginationLimit int
	ServerPort         string
	ServerTimeout      time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		DBURL:              getEnv("DBURL", "postgres://user:password@localhost:5432/ssubench?sslmode=disable"),
		JWTSecret:          getEnv("JWT_SECRET", "very_secret_key"),
		JWTTTL:             time.Hour * 24,
		PasswordCost:       10,
		StartingBalance:    0,
		MaxPaginationLimit: 100,
		ServerPort:         getEnv("PORT", "8080"),
		ServerTimeout:      15 * time.Second,
		ReadTimeout:        15 * time.Second,
		WriteTimeout:       15 * time.Second,
		IdleTimeout:        60 * time.Second,
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)

	if exists {
		return value
	}

	return defaultValue
}
