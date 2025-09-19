package util

import (
	"os"

	"github.com/joho/godotenv"
)

// Project configuration
type Config struct {
	// Server config
	BaseURL string

	// Database config
	DBConn string
}

// Load config from enviroment
func LoadConfig(path string) *Config {
	godotenv.Load(path)
	return &Config{
		BaseURL: os.Getenv("BASE_URL"),
		DBConn:  os.Getenv("DB_CONN"),
	}
}
