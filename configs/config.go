package configs

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"url/short/pkg/errors"
)

// Config represents the application configuration
type Config struct {
	DB   DBConfig
	Auth AuthConfig
}

// DBConfig holds database configuration
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// DSN returns the DSN string for the database connection
func (c *DBConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Secret string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Set default values
	config := &Config{
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", ""),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Auth: AuthConfig{
			Secret: getEnv("JWT_SECRET", "your_secure_jwt_secret_here"),
		},
	}

	if err := config.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	return config
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Check database configuration
	if c.DB.Host == "" {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "DB_HOST is required")
	}
	if c.DB.Port == "" {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "DB_PORT is required")
	}
	if c.DB.User == "" {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "DB_USER is required")
	}
	if c.DB.Name == "" {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "DB_NAME is required")
	}

	if c.Auth.Secret == "" {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "Auth secret is required")
	}

	if len(c.Auth.Secret) < 32 {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "Auth secret must be at least 32 characters long")
	}

	return nil
}
