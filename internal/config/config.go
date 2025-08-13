package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port     int              `json:"port"`
	Database *DatabaseConfig  `json:"database"`
	JWT      *JWTConfig       `json:"jwt"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
	SSLMode  string `json:"ssl_mode"`
}

type JWTConfig struct {
	Secret string `json:"secret"`
	Expiry string `json:"expiry"`
}

func Load() (*Config, error) {
	config := &Config{
		Port: getEnvAsInt("PORT", 8080),
		Database: &DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "bankdb"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: &JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
			Expiry: getEnv("JWT_EXPIRY", "24h"),
		},
	}

	return config, nil
}

func (db *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		db.User, db.Password, db.Host, db.Port, db.Name, db.SSLMode)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}