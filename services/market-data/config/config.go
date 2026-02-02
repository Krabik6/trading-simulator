package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Service ServiceConfig
	Client  ClientConfig
	Kafka   KafkaConfig
}

type ServiceConfig struct {
	Name           string
	HTTPPort       int
	LogLevel       string
	MetricsEnabled bool
}

type ClientConfig struct {
	Type    string
	Symbols []string
}

type KafkaConfig struct {
	Brokers        []string
	Topic          string
	BatchSize      int
	BatchTimeout   int // milliseconds
	ConnectRetries int
	RetryInterval  int // seconds
}

func Load() (*Config, error) {
	cfg := &Config{
		Service: ServiceConfig{
			Name:           getEnv("SERVICE_NAME", "market-data"),
			HTTPPort:       getEnvInt("HTTP_PORT", 8080),
			LogLevel:       getEnv("LOG_LEVEL", "info"),
			MetricsEnabled: getEnvBool("METRICS_ENABLED", true),
		},
		Client: ClientConfig{
			Type:    getEnv("CLIENT_TYPE", "mock"),
			Symbols: getEnvSlice("SYMBOLS", []string{"BTCUSDT", "ETHUSDT", "SOLUSDT"}),
		},
		Kafka: KafkaConfig{
			Brokers:        getEnvSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
			Topic:          getEnv("KAFKA_TOPIC", "crypto-prices"),
			BatchSize:      getEnvInt("KAFKA_BATCH_SIZE", 100),
			BatchTimeout:   getEnvInt("KAFKA_BATCH_TIMEOUT_MS", 100),
			ConnectRetries: getEnvInt("KAFKA_CONNECT_RETRIES", 10),
			RetryInterval:  getEnvInt("KAFKA_RETRY_INTERVAL_SEC", 3),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	var errs []string

	// Service validation
	if c.Service.HTTPPort < 1 || c.Service.HTTPPort > 65535 {
		errs = append(errs, fmt.Sprintf("invalid HTTP_PORT: %d", c.Service.HTTPPort))
	}

	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[strings.ToLower(c.Service.LogLevel)] {
		errs = append(errs, fmt.Sprintf("invalid LOG_LEVEL: %s (must be debug/info/warn/error)", c.Service.LogLevel))
	}

	// Client validation
	validClientTypes := map[string]bool{"mock": true, "binance": true}
	if !validClientTypes[c.Client.Type] {
		errs = append(errs, fmt.Sprintf("invalid CLIENT_TYPE: %s (must be mock/binance)", c.Client.Type))
	}

	if len(c.Client.Symbols) == 0 {
		errs = append(errs, "SYMBOLS cannot be empty")
	}

	// Kafka validation
	if len(c.Kafka.Brokers) == 0 {
		errs = append(errs, "KAFKA_BROKERS cannot be empty")
	}

	if c.Kafka.Topic == "" {
		errs = append(errs, "KAFKA_TOPIC cannot be empty")
	}

	if c.Kafka.BatchSize < 1 {
		errs = append(errs, fmt.Sprintf("invalid KAFKA_BATCH_SIZE: %d", c.Kafka.BatchSize))
	}

	if len(errs) > 0 {
		return errors.New("config validation failed: " + strings.Join(errs, "; "))
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
