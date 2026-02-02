package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Service  ServiceConfig
	Database DatabaseConfig
	Kafka    KafkaConfig
	JWT      JWTConfig
	Trading  TradingConfig
}

type ServiceConfig struct {
	Name           string
	HTTPPort       int
	LogLevel       string
	MetricsEnabled bool
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type KafkaConfig struct {
	Brokers        []string
	PricesTopic    string
	TradesTopic    string
	ConsumerGroup  string
	ConnectRetries int
	RetryInterval  int
}

type JWTConfig struct {
	Secret      string
	ExpiryHours int
}

type TradingConfig struct {
	MaxLeverage      int
	InitialBalance   float64
	SupportedSymbols []string
	MaintenanceRate  float64 // maintenance margin rate (e.g., 0.005 = 0.5%)
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

func Load() (*Config, error) {
	cfg := &Config{
		Service: ServiceConfig{
			Name:           getEnv("SERVICE_NAME", "trading"),
			HTTPPort:       getEnvInt("HTTP_PORT", 8081),
			LogLevel:       getEnv("LOG_LEVEL", "info"),
			MetricsEnabled: getEnvBool("METRICS_ENABLED", true),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "trading"),
			Password:        getEnv("DB_PASSWORD", "trading"),
			Name:            getEnv("DB_NAME", "trading"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: time.Duration(getEnvInt("DB_CONN_MAX_LIFETIME_MIN", 5)) * time.Minute,
		},
		Kafka: KafkaConfig{
			Brokers:        getEnvSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
			PricesTopic:    getEnv("KAFKA_PRICES_TOPIC", "crypto-prices"),
			TradesTopic:    getEnv("KAFKA_TRADES_TOPIC", "trades"),
			ConsumerGroup:  getEnv("KAFKA_CONSUMER_GROUP", "trading-service"),
			ConnectRetries: getEnvInt("KAFKA_CONNECT_RETRIES", 10),
			RetryInterval:  getEnvInt("KAFKA_RETRY_INTERVAL_SEC", 3),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			ExpiryHours: getEnvInt("JWT_EXPIRY_HOURS", 24),
		},
		Trading: TradingConfig{
			MaxLeverage:      getEnvInt("MAX_LEVERAGE", 100),
			InitialBalance:   getEnvFloat("INITIAL_BALANCE", 10000),
			SupportedSymbols: getEnvSlice("SUPPORTED_SYMBOLS", []string{"BTCUSDT", "ETHUSDT", "SOLUSDT"}),
			MaintenanceRate:  getEnvFloat("MAINTENANCE_RATE", 0.005),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	var errs []string

	if c.Service.HTTPPort < 1 || c.Service.HTTPPort > 65535 {
		errs = append(errs, fmt.Sprintf("invalid HTTP_PORT: %d", c.Service.HTTPPort))
	}

	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[strings.ToLower(c.Service.LogLevel)] {
		errs = append(errs, fmt.Sprintf("invalid LOG_LEVEL: %s", c.Service.LogLevel))
	}

	if c.Database.Host == "" {
		errs = append(errs, "DB_HOST cannot be empty")
	}

	if len(c.Kafka.Brokers) == 0 {
		errs = append(errs, "KAFKA_BROKERS cannot be empty")
	}

	if c.JWT.Secret == "" {
		errs = append(errs, "JWT_SECRET cannot be empty")
	}

	if c.Trading.MaxLeverage < 1 || c.Trading.MaxLeverage > 125 {
		errs = append(errs, fmt.Sprintf("invalid MAX_LEVERAGE: %d (must be 1-125)", c.Trading.MaxLeverage))
	}

	if c.Trading.InitialBalance <= 0 {
		errs = append(errs, "INITIAL_BALANCE must be positive")
	}

	if len(c.Trading.SupportedSymbols) == 0 {
		errs = append(errs, "SUPPORTED_SYMBOLS cannot be empty")
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

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
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
