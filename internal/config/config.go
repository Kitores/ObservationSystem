package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	PostgresDB       string
	PostgresUser     string
	PostgresPassword string
	PostgresHost     string
	PostgresPort     string
	PostgresSSLMode  string

	JWTToken string
}

func InitConfig() (cfg *Config, err error) {
	configPath := ".env"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("CONFIG_PATH file does not exist: %s", configPath)
	}

	err = godotenv.Load(configPath)
	if err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
		os.Exit(1)
	}
	cfg = new(Config)
	cfg.JWTToken = os.Getenv("JWT_TOKEN")
	cfg.PostgresSSLMode = os.Getenv("POSTGRES_SSL_MODE")
	cfg.PostgresUser = os.Getenv("POSTGRES_USER")
	cfg.PostgresPassword = os.Getenv("POSTGRES_PASSWORD")
	cfg.PostgresHost = os.Getenv("POSTGRES_HOST")
	cfg.PostgresPort = os.Getenv("POSTGRES_PORT")
	cfg.PostgresDB = os.Getenv("POSTGRES_DB")

	return cfg, err
}
