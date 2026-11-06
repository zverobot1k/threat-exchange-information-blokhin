package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Address     string
}

func Load() (Config, error) {
	_ = godotenv.Load()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	address := os.Getenv("SERVER_ADDRESS")
	if address == "" {
		address = ":8080"
	}
	return Config{DatabaseURL: databaseURL, Address: address}, nil
}
