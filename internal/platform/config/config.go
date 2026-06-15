package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

const masterKeyByteLength = 32

type Config struct {
	Server     ServerConfig   `mapstructure:"server"`
	Worker     WorkerConfig   `mapstructure:"worker"`
	Gateway    GatewayConfig  `mapstructure:"gateway"`
	AsynqRedis RedisConfig    `mapstructure:"asynq_redis"`
	CacheRedis RedisConfig    `mapstructure:"cache_redis"`
	Security   SecurityConfig `mapstructure:"security"`
}

type ServerConfig struct {
	Port     int            `mapstructure:"port"`
	Database DatabaseConfig `mapstructure:"database"`
}

type WorkerConfig struct {
	HealthPort int `mapstructure:"health_port"`
}

type GatewayConfig struct {
	Port     int            `mapstructure:"port"`
	Database DatabaseConfig `mapstructure:"database"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

type RedisConfig struct {
	Addr string `mapstructure:"addr"`
	DB   int    `mapstructure:"db"`
}

type SecurityConfig struct {
	InternalServiceToken string `mapstructure:"internal_service_token"`
	MasterKeyBase64      string `mapstructure:"master_key_base64"`
}

var environmentBindings = map[string]string{
	"server.port":                     "OPSWEAVER_SERVER_PORT",
	"worker.health_port":              "OPSWEAVER_WORKER_HEALTH_PORT",
	"gateway.port":                    "OPSWEAVER_GATEWAY_PORT",
	"server.database.dsn":             "OPSWEAVER_SERVER_DATABASE_DSN",
	"gateway.database.dsn":            "OPSWEAVER_GATEWAY_DATABASE_DSN",
	"asynq_redis.addr":                "ASYNQ_REDIS_ADDR",
	"asynq_redis.db":                  "ASYNQ_REDIS_DB",
	"cache_redis.addr":                "CACHE_REDIS_ADDR",
	"cache_redis.db":                  "CACHE_REDIS_DB",
	"security.internal_service_token": "INTERNAL_SERVICE_TOKEN",
	"security.master_key_base64":      "MASTER_KEY_BASE64",
}

func Load(path string) (Config, error) {
	v := viper.New()
	v.SetConfigFile(path)

	for key, environmentVariable := range environmentBindings {
		if err := v.BindEnv(key, environmentVariable); err != nil {
			return Config{}, fmt.Errorf("bind environment variable %s: %w", environmentVariable, err)
		}
	}

	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return Config{}, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

func (cfg Config) validate() error {
	if err := validatePort("server.port", cfg.Server.Port); err != nil {
		return err
	}
	if err := validatePort("worker.health_port", cfg.Worker.HealthPort); err != nil {
		return err
	}
	if err := validatePort("gateway.port", cfg.Gateway.Port); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Server.Database.DSN) == "" {
		return errors.New("server.database.dsn is required")
	}
	if strings.TrimSpace(cfg.Gateway.Database.DSN) == "" {
		return errors.New("gateway.database.dsn is required")
	}
	if cfg.Server.Database.DSN == cfg.Gateway.Database.DSN {
		return errors.New("server and gateway database DSNs must be different")
	}
	if strings.TrimSpace(cfg.AsynqRedis.Addr) == "" {
		return errors.New("asynq_redis.addr is required")
	}
	if cfg.AsynqRedis.DB != 0 {
		return errors.New("asynq_redis.db must be 0")
	}
	if strings.TrimSpace(cfg.CacheRedis.Addr) == "" {
		return errors.New("cache_redis.addr is required")
	}
	if cfg.CacheRedis.DB != 1 {
		return errors.New("cache_redis.db must be 1")
	}
	if strings.TrimSpace(cfg.Security.InternalServiceToken) == "" {
		return errors.New("security.internal_service_token is required")
	}
	if strings.TrimSpace(cfg.Security.MasterKeyBase64) == "" {
		return errors.New("security.master_key_base64 is required")
	}
	decoded, err := base64.StdEncoding.DecodeString(cfg.Security.MasterKeyBase64)
	if err != nil {
		return errors.New("security.master_key_base64 must be valid base64")
	}
	if len(decoded) != masterKeyByteLength {
		return fmt.Errorf("security.master_key_base64 must decode to a %d-byte key", masterKeyByteLength)
	}

	return nil
}

func validatePort(name string, port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("%s must be between 1 and 65535", name)
	}
	return nil
}
