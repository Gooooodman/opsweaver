package config_test

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Gooooodman/opsweaver/internal/platform/config"
)

func TestLoadValidYAML(t *testing.T) {
	path := writeConfig(t, validConfigYAML)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.Worker.HealthPort != 8081 {
		t.Errorf("Worker.HealthPort = %d, want 8081", cfg.Worker.HealthPort)
	}
	if cfg.Gateway.Port != 8082 {
		t.Errorf("Gateway.Port = %d, want 8082", cfg.Gateway.Port)
	}
	if cfg.Server.Database.DSN != "postgres://localhost/opsweaver_server_db" {
		t.Errorf("Server.Database.DSN = %q, want server DSN", cfg.Server.Database.DSN)
	}
	if cfg.Gateway.Database.DSN != "postgres://localhost/opsweaver_gateway_db" {
		t.Errorf("Gateway.Database.DSN = %q, want gateway DSN", cfg.Gateway.Database.DSN)
	}
	if cfg.AsynqRedis.DB != 0 {
		t.Errorf("AsynqRedis.DB = %d, want 0", cfg.AsynqRedis.DB)
	}
	if cfg.CacheRedis.DB != 1 {
		t.Errorf("CacheRedis.DB = %d, want 1", cfg.CacheRedis.DB)
	}
}

func TestLoadRejectsMissingRequiredValue(t *testing.T) {
	path := writeConfig(t, strings.Replace(
		validConfigYAML,
		"  internal_service_token: test-internal-token\n",
		"",
		1,
	))
	t.Setenv("INTERNAL_SERVICE_TOKEN", "")

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want missing required value error")
	}
	if !strings.Contains(err.Error(), "security.internal_service_token is required") {
		t.Fatalf("Load() error = %q, want required field message", err)
	}
}

func TestLoadAppliesEnvironmentOverrides(t *testing.T) {
	path := writeConfig(t, validConfigYAML)
	t.Setenv("OPSWEAVER_SERVER_PORT", "9090")
	t.Setenv("OPSWEAVER_WORKER_HEALTH_PORT", "9091")
	t.Setenv("OPSWEAVER_GATEWAY_PORT", "9092")
	t.Setenv("OPSWEAVER_SERVER_DATABASE_DSN", "postgres://override/opsweaver_server_db")
	t.Setenv("OPSWEAVER_GATEWAY_DATABASE_DSN", "postgres://override/opsweaver_gateway_db")
	t.Setenv("ASYNQ_REDIS_ADDR", "asynq.example:6379")
	t.Setenv("ASYNQ_REDIS_DB", "0")
	t.Setenv("CACHE_REDIS_ADDR", "cache.example:6379")
	t.Setenv("CACHE_REDIS_DB", "1")
	t.Setenv("INTERNAL_SERVICE_TOKEN", "override-token")
	t.Setenv("MASTER_KEY_BASE64", base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")))

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %d, want 9090", cfg.Server.Port)
	}
	if cfg.Worker.HealthPort != 9091 {
		t.Errorf("Worker.HealthPort = %d, want 9091", cfg.Worker.HealthPort)
	}
	if cfg.Gateway.Port != 9092 {
		t.Errorf("Gateway.Port = %d, want 9092", cfg.Gateway.Port)
	}
	if cfg.Server.Database.DSN != "postgres://override/opsweaver_server_db" {
		t.Errorf("Server.Database.DSN = %q, want overridden DSN", cfg.Server.Database.DSN)
	}
	if cfg.Gateway.Database.DSN != "postgres://override/opsweaver_gateway_db" {
		t.Errorf("Gateway.Database.DSN = %q, want overridden DSN", cfg.Gateway.Database.DSN)
	}
	if cfg.AsynqRedis.Addr != "asynq.example:6379" || cfg.AsynqRedis.DB != 0 {
		t.Errorf("AsynqRedis = %+v, want overridden address and DB 0", cfg.AsynqRedis)
	}
	if cfg.CacheRedis.Addr != "cache.example:6379" || cfg.CacheRedis.DB != 1 {
		t.Errorf("CacheRedis = %+v, want overridden address and DB 1", cfg.CacheRedis)
	}
	if cfg.Security.InternalServiceToken != "override-token" {
		t.Errorf("Security.InternalServiceToken = %q, want overridden token", cfg.Security.InternalServiceToken)
	}
	if cfg.Security.MasterKeyBase64 != base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")) {
		t.Errorf("Security.MasterKeyBase64 was not overridden")
	}
}

func TestLoadRejectsSensitiveValuesFromYAMLWithoutEnvironment(t *testing.T) {
	path := writeConfig(t, validConfigYAML)
	t.Setenv("INTERNAL_SERVICE_TOKEN", "")

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want environment-only secret error")
	}
	if !strings.Contains(err.Error(), "security.internal_service_token is required") {
		t.Fatalf("Load() error = %q, want required token message", err)
	}
}

func TestLoadRejectsExplicitEmptyEnvironmentOverride(t *testing.T) {
	path := writeConfig(t, validConfigYAML)
	t.Setenv("OPSWEAVER_SERVER_DATABASE_DSN", "")

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want empty DSN error")
	}
	if !strings.Contains(err.Error(), "server.database.dsn is required") {
		t.Fatalf("Load() error = %q, want required DSN message", err)
	}
}

func TestLoadRejectsUnknownYAMLField(t *testing.T) {
	path := writeConfig(t, strings.Replace(
		validConfigYAML,
		"asynq_redis:\n  addr:",
		"asynq_redis:\n  typo: true\n  addr:",
		1,
	))

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want unknown field error")
	}
	if !strings.Contains(err.Error(), "typo") {
		t.Fatalf("Load() error = %q, want unknown field name", err)
	}
}

func TestLoadRejectsNonNumericEnvironmentValue(t *testing.T) {
	path := writeConfig(t, validConfigYAML)
	t.Setenv("OPSWEAVER_SERVER_PORT", "not-a-number")

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want numeric conversion error")
	}
	if !strings.Contains(err.Error(), "server.port") {
		t.Fatalf("Load() error = %q, want server.port context", err)
	}
}

func TestLoadReadErrorIncludesConfigPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.yaml")

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want read error")
	}
	if !strings.Contains(err.Error(), path) {
		t.Fatalf("Load() error = %q, want config path %q", err, path)
	}
}

func TestLoadMalformedYAMLErrorIncludesConfigPath(t *testing.T) {
	path := writeConfig(t, "server: [\n")

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want parse error")
	}
	if !strings.Contains(err.Error(), path) {
		t.Fatalf("Load() error = %q, want config path %q", err, path)
	}
}

func TestLoadRejectsPortOutsideValidRange(t *testing.T) {
	path := writeConfig(t, validConfigYAML)
	t.Setenv("OPSWEAVER_GATEWAY_PORT", "70000")

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want invalid port error")
	}
	if !strings.Contains(err.Error(), "gateway.port must be between 1 and 65535") {
		t.Fatalf("Load() error = %q, want port range message", err)
	}
}

func TestLoadRejectsSharedDatabaseDSN(t *testing.T) {
	path := writeConfig(t, validConfigYAML)
	t.Setenv("OPSWEAVER_GATEWAY_DATABASE_DSN", "postgres://localhost/opsweaver_server_db")

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want independent database DSN error")
	}
	if !strings.Contains(err.Error(), "server and gateway database DSNs must be different") {
		t.Fatalf("Load() error = %q, want independent database DSN message", err)
	}
}

func TestLoadEnforcesRedisDatabaseAssignments(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		want     string
	}{
		{
			name:     "asynq",
			envKey:   "ASYNQ_REDIS_DB",
			envValue: "1",
			want:     "asynq_redis.db must be 0",
		},
		{
			name:     "cache",
			envKey:   "CACHE_REDIS_DB",
			envValue: "0",
			want:     "cache_redis.db must be 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeConfig(t, validConfigYAML)
			t.Setenv(tt.envKey, tt.envValue)

			_, err := config.Load(path)
			if err == nil {
				t.Fatal("Load() error = nil, want Redis database error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Load() error = %q, want %q", err, tt.want)
			}
		})
	}
}

func TestLoadRejectsInvalidMasterKey(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     string
	}{
		{
			name:     "not base64",
			envValue: "!!!not-base64!!!",
			want:     "must be valid base64",
		},
		{
			name:     "wrong length",
			envValue: base64.StdEncoding.EncodeToString([]byte("short-key")),
			want:     "must decode to a 32-byte key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeConfig(t, validConfigYAML)
			t.Setenv("MASTER_KEY_BASE64", tt.envValue)

			_, err := config.Load(path)
			if err == nil {
				t.Fatal("Load() error = nil, want master key error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Load() error = %q, want %q", err, tt.want)
			}
		})
	}
}

func TestLoadErrorOmitsSecretValues(t *testing.T) {
	const sentinelToken = "sentinel-secret-do-not-leak-9f3a"
	const sentinelKey = "sentinel-key-do-not-leak-7b1d-not-base64!!!"

	path := writeConfig(t, validConfigYAML)
	t.Setenv("INTERNAL_SERVICE_TOKEN", sentinelToken)
	t.Setenv("MASTER_KEY_BASE64", sentinelKey)

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want master key error")
	}
	message := err.Error()
	if strings.Contains(message, sentinelToken) {
		t.Errorf("Load() error leaked internal service token: %q", message)
	}
	if strings.Contains(message, sentinelKey) {
		t.Errorf("Load() error leaked master key: %q", message)
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()

	for _, key := range []string{
		"OPSWEAVER_SERVER_PORT",
		"OPSWEAVER_WORKER_HEALTH_PORT",
		"OPSWEAVER_GATEWAY_PORT",
		"OPSWEAVER_SERVER_DATABASE_DSN",
		"OPSWEAVER_GATEWAY_DATABASE_DSN",
		"ASYNQ_REDIS_ADDR",
		"ASYNQ_REDIS_DB",
		"CACHE_REDIS_ADDR",
		"CACHE_REDIS_DB",
	} {
		unsetEnv(t, key)
	}
	t.Setenv("INTERNAL_SERVICE_TOKEN", "test-internal-token")
	t.Setenv("MASTER_KEY_BASE64", base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")))

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	value, existed := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("Unsetenv(%q) error = %v", key, err)
	}
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, value)
			return
		}
		_ = os.Unsetenv(key)
	})
}

const validConfigYAML = `
server:
  port: 8080
  database:
    dsn: postgres://localhost/opsweaver_server_db
worker:
  health_port: 8081
gateway:
  port: 8082
  database:
    dsn: postgres://localhost/opsweaver_gateway_db
asynq_redis:
  addr: localhost:6379
  db: 0
cache_redis:
  addr: localhost:6379
  db: 1
security:
  internal_service_token: test-internal-token
  master_key_base64: MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=
`
