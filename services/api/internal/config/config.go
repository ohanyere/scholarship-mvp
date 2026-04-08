package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv              string
	LogLevel            string
	HTTPPort            int
	RequestTimeout      time.Duration
	ShutdownTimeout     time.Duration
	PostgresHost        string
	PostgresPort        int
	PostgresDB          string
	PostgresUser        string
	PostgresPassword    string
	PostgresSSLMode     string
	RedisHost           string
	RedisPort           int
	RedisPassword       string
	RedisDB             int
	JWTSecret           string
	JWTAccessTokenTTL   time.Duration
	ScholarshipCacheTTL time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:              getEnv("APP_ENV", "local"),
		LogLevel:            getEnv("LOG_LEVEL", "info"),
		HTTPPort:            getEnvInt("API_PORT", 8080),
		RequestTimeout:      getEnvDuration("REQUEST_TIMEOUT", 15*time.Second),
		ShutdownTimeout:     getEnvDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
		PostgresHost:        getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:        getEnvInt("POSTGRES_PORT", 5432),
		PostgresDB:          getEnv("POSTGRES_DB", "scholarship_platform"),
		PostgresUser:        getEnv("POSTGRES_USER", "scholarship_platform"),
		PostgresPassword:    getEnv("POSTGRES_PASSWORD", "change-me"),
		PostgresSSLMode:     getEnv("POSTGRES_SSLMODE", "disable"),
		RedisHost:           getEnv("REDIS_HOST", "localhost"),
		RedisPort:           getEnvInt("REDIS_PORT", 6379),
		RedisPassword:       getEnv("REDIS_PASSWORD", ""),
		RedisDB:             getEnvInt("REDIS_DB", 0),
		JWTSecret:           getEnv("JWT_SECRET", "change-me"),
		JWTAccessTokenTTL:   getEnvDuration("JWT_ACCESS_TOKEN_TTL", 15*time.Minute),
		ScholarshipCacheTTL: getEnvDuration("CACHE_TTL_SCHOLARSHIPS", 5*time.Minute),
	}

	if cfg.HTTPPort <= 0 {
		return Config{}, fmt.Errorf("api port must be positive")
	}

	return cfg, nil
}

func (c Config) PostgresDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		c.PostgresHost,
		c.PostgresPort,
		c.PostgresDB,
		c.PostgresUser,
		c.PostgresPassword,
		c.PostgresSSLMode,
	)
}

func (c Config) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.RedisHost, c.RedisPort)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func getEnvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}

	return value
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}

	return value
}
