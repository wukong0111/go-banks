package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port     int             `json:"port"`
	Database *DatabaseConfig `json:"database"`
	JWT      *JWTConfig      `json:"jwt"`
	Logger   *LoggerConfig   `json:"logger"`
	APIKey   string          `json:"api_key"`
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

type LoggerConfig struct {
	Level       string   `json:"level"`
	Outputs     []string `json:"outputs"`
	JSONFile    string   `json:"json_file,omitempty"`
	TextFile    string   `json:"text_file,omitempty"`
	AddSource   bool     `json:"add_source"`
	MaxFileSize int64    `json:"max_file_size"`
}

func Load() (*Config, error) {
	config := &Config{
		Port:   getEnvAsInt("PORT", 8080),
		APIKey: getEnv("API_KEY", "dev-api-key"),
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
		Logger: &LoggerConfig{
			Level:       getEnv("LOG_LEVEL", "info"),
			Outputs:     parseLogOutputs(getEnv("LOG_OUTPUTS", "console")),
			JSONFile:    getEnv("LOG_JSON_FILE", ""),
			TextFile:    getEnv("LOG_TEXT_FILE", ""),
			AddSource:   getEnvAsBool("LOG_ADD_SOURCE", false),
			MaxFileSize: getEnvAsInt64("LOG_MAX_FILE_SIZE", 100*1024*1024), // 100MB
		},
	}

	slog.Info("configuration loaded successfully",
		"port", config.Port,
		"db_host", config.Database.Host,
		"db_port", config.Database.Port,
		"db_name", config.Database.Name,
		"jwt_expiry", config.JWT.Expiry,
		"log_level", config.Logger.Level,
		"log_outputs", config.Logger.Outputs,
	)

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
		intValue, err := strconv.Atoi(value)
		if err == nil {
			return intValue
		}
		slog.Warn("invalid environment variable, using default",
			"key", key,
			"value", value,
			"default", defaultValue,
			"error", err.Error(),
		)
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			return intValue
		}
		slog.Warn("invalid environment variable, using default",
			"key", key,
			"value", value,
			"default", defaultValue,
			"error", err.Error(),
		)
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		boolValue, err := strconv.ParseBool(value)
		if err == nil {
			return boolValue
		}
		slog.Warn("invalid environment variable, using default",
			"key", key,
			"value", value,
			"default", defaultValue,
			"error", err.Error(),
		)
	}
	return defaultValue
}

func parseLogOutputs(outputsStr string) []string {
	if outputsStr == "" {
		return []string{"console"}
	}

	outputs := strings.Split(outputsStr, ",")
	for i, output := range outputs {
		outputs[i] = strings.TrimSpace(output)
	}
	return outputs
}
