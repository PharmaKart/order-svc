package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	DBConnString      string
	ProductServiceURL string
	PaymentServiceURL string
}

// LoadConfig loads configuration from environment variables or a .env file.
func LoadConfig() *Config {
	// Load environment variables from .env file (if it exists)
	if err := godotenv.Overload(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	return &Config{
		Port:              getEnv("PORT", "50053"),
		DBConnString:      getDBConnString(),
		ProductServiceURL: getEnv("PRODUCT_SERVICE_URL", "localhost:50052"),
		PaymentServiceURL: getEnv("PAYMENT_SERVICE_URL", "localhost:50054"),
	}
}

// getDBConnString constructs the PostgreSQL connection string.
func getDBConnString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=prefer",
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_NAME", "pharmakartdb"),
	)
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
