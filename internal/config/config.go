package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Binance  BinanceConfig  `mapstructure:"binance"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Sync     SyncConfig     `mapstructure:"sync"`
	Stream   StreamConfig   `mapstructure:"stream"`
}

// AppConfig holds application-level configuration
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Environment string `mapstructure:"environment"`
	LogLevel    string `mapstructure:"log_level"`
}

// BinanceConfig holds Binance API configuration
type BinanceConfig struct {
	APIURL         string   `mapstructure:"api_url"`
	WSURL          string   `mapstructure:"ws_url"`
	RestRateLimit  int      `mapstructure:"rest_rate_limit"`
	KlineIntervals []string `mapstructure:"kline_intervals"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host                  string `mapstructure:"host"`
	Port                  int    `mapstructure:"port"`
	User                  string `mapstructure:"user"`
	Password              string `mapstructure:"password"`
	Database              string `mapstructure:"database"`
	SSLMode               string `mapstructure:"ssl_mode"`
	MaxConnections        int    `mapstructure:"max_connections"`
	MaxIdleConnections    int    `mapstructure:"max_idle_connections"`
	ConnectionMaxLifetime int    `mapstructure:"connection_max_lifetime"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Password    string `mapstructure:"password"`
	DB          int    `mapstructure:"db"`
	PoolSize    int    `mapstructure:"pool_size"`
	LiveDataTTL int    `mapstructure:"live_data_ttl"`
}

// SyncConfig holds data synchronization configuration
type SyncConfig struct {
	Enabled      bool `mapstructure:"enabled"`
	MaxSyncHours int  `mapstructure:"max_sync_hours"`
	BatchSize    int  `mapstructure:"batch_size"`
	Workers      int  `mapstructure:"workers"`
}

// StreamConfig holds WebSocket streaming configuration
type StreamConfig struct {
	ReconnectDelay       int `mapstructure:"reconnect_delay"`
	MaxReconnectAttempts int `mapstructure:"max_reconnect_attempts"`
	PingInterval         int `mapstructure:"ping_interval"`
	ChannelBufferSize    int `mapstructure:"channel_buffer_size"`
}

// Load reads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Set config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath(".")
	}

	// Enable environment variable override
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read configuration
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal configuration
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "binance-live-collector")
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.log_level", "info")

	v.SetDefault("binance.api_url", "https://api.binance.com")
	v.SetDefault("binance.ws_url", "wss://stream.binance.com:9443")
	v.SetDefault("binance.rest_rate_limit", 1200)
	v.SetDefault("binance.kline_intervals", []string{"1m", "5m", "1h", "1d"})

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.database", "binance_data")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_connections", 25)
	v.SetDefault("database.max_idle_connections", 5)
	v.SetDefault("database.connection_max_lifetime", 300)

	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.live_data_ttl", 60)

	v.SetDefault("sync.enabled", true)
	v.SetDefault("sync.max_sync_hours", 24)
	v.SetDefault("sync.batch_size", 1000)
	v.SetDefault("sync.workers", 5)

	v.SetDefault("stream.reconnect_delay", 5)
	v.SetDefault("stream.max_reconnect_attempts", 10)
	v.SetDefault("stream.ping_interval", 30)
	v.SetDefault("stream.channel_buffer_size", 1000)
}

// GetDSN returns the PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

// GetRedisAddr returns the Redis address
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
