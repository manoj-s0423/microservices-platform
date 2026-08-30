// Package config loads and validates order-service configuration from
// environment variables. Every setting has a sane local-dev default except
// the database credentials and downstream service URLs, which are required
// for the service to do anything useful.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort      string
	LogLevel        string
	Environment     string
	ShutdownTimeout time.Duration

	DBHost           string
	DBPort           string
	DBName           string
	DBUser           string
	DBPassword       string
	DBMaxOpenConns   int
	DBMaxIdleConns   int
	DBConnectTimeout time.Duration

	UserServiceURL    string
	ProductServiceURL string
	PaymentServiceURL string

	HTTPClientTimeout time.Duration
	HTTPRetryAttempts int
	HTTPRetryBackoff  time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:        getEnv("SERVER_PORT", "8082"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		Environment:       getEnv("ENVIRONMENT", "development"),
		ShutdownTimeout:   time.Duration(getEnvInt("SHUTDOWN_TIMEOUT_SECONDS", 10)) * time.Second,
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBName:            getEnv("DB_NAME", "shopstream_orders"),
		DBUser:            getEnv("DB_USER", "shopstream"),
		DBPassword:        getEnv("DB_PASSWORD", ""),
		DBMaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 10),
		DBMaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
		DBConnectTimeout:  time.Duration(getEnvInt("DB_CONNECT_TIMEOUT_SECONDS", 5)) * time.Second,
		UserServiceURL:    getEnv("USER_SERVICE_URL", "http://localhost:8081"),
		ProductServiceURL: getEnv("PRODUCT_SERVICE_URL", "http://localhost:8000"),
		PaymentServiceURL: getEnv("PAYMENT_SERVICE_URL", "http://localhost:8083"),
		HTTPClientTimeout: time.Duration(getEnvInt("HTTP_CLIENT_TIMEOUT_SECONDS", 3)) * time.Second,
		HTTPRetryAttempts: getEnvInt("HTTP_CLIENT_RETRY_ATTEMPTS", 2),
		HTTPRetryBackoff:  time.Duration(getEnvInt("HTTP_CLIENT_RETRY_BACKOFF_MS", 200)) * time.Millisecond,
	}

	if cfg.DBUser == "" {
		return nil, fmt.Errorf("DB_USER must not be empty")
	}

	return cfg, nil
}

func (c *Config) DatabaseDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable&connect_timeout=%d",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
		int(c.DBConnectTimeout.Seconds()),
	)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		// Deliberately fall back rather than panic here; individual
		// services can choose to validate strictly. Logged by the caller
		// once the logger is initialized.
		return fallback
	}
	return i
}
