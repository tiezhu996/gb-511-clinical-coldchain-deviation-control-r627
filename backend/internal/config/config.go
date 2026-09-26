package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the complete runtime contract. Every value can be overridden by
// environment variables so the same binary works locally and in Compose.
type Config struct {
	AppName           string
	Environment       string
	Port              string
	DatabaseDriver    string
	DatabaseDSN       string
	RedisAddr         string
	RedisPassword     string
	MinIOEndpoint     string
	MinIOAccessKey    string
	MinIOSecretKey    string
	MinIOBucket       string
	MinIOUseSSL       bool
	JWTSecret         string
	TokenTTL          time.Duration
	RequestLimit      int
	TrustedProxies    []string
	StartupTimeout    time.Duration
	ShutdownTimeout   time.Duration
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

// PublicConfig is safe to expose from the authenticated runtime endpoint.
// Connection strings, passwords and signing keys are intentionally omitted.
type PublicConfig struct {
	AppName         string `json:"appName"`
	Environment     string `json:"environment"`
	DatabaseDriver  string `json:"databaseDriver"`
	RedisEnabled    bool   `json:"redisEnabled"`
	TokenTTLSeconds int64  `json:"tokenTtlSeconds"`
	RequestLimit    int    `json:"requestLimit"`
	ShutdownSeconds int64  `json:"shutdownSeconds"`
}

func Load() (Config, error) {
	cfg := Config{
		AppName:           env("APP_NAME", "clinical-coldchain-deviation-control"),
		Environment:       env("APP_ENV", "development"),
		Port:              env("PORT", "8080"),
		DatabaseDriver:    strings.ToLower(env("DATABASE_DRIVER", "postgres")),
		DatabaseDSN:       env("DATABASE_DSN", "app.db"),
		RedisAddr:         env("REDIS_ADDR", ""),
		RedisPassword:     env("REDIS_PASSWORD", ""),
		MinIOEndpoint:     env("MINIO_ENDPOINT", "127.0.0.1:22511"),
		MinIOAccessKey:    env("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey:    env("MINIO_SECRET_KEY", "minioadmin-change-me"),
		MinIOBucket:       env("MINIO_BUCKET", "clinical-evidence"),
		MinIOUseSSL:       boolEnv("MINIO_USE_SSL", false),
		JWTSecret:         env("JWT_SECRET", "development-only-change-me"),
		TokenTTL:          durationEnv("TOKEN_TTL", 8*time.Hour),
		RequestLimit:      intEnv("REQUEST_LIMIT", 180),
		TrustedProxies:    splitCSV(env("TRUSTED_PROXIES", "")),
		StartupTimeout:    durationEnv("STARTUP_TIMEOUT", 35*time.Second),
		ShutdownTimeout:   durationEnv("SHUTDOWN_TIMEOUT", 10*time.Second),
		ReadHeaderTimeout: durationEnv("READ_HEADER_TIMEOUT", 5*time.Second),
		ReadTimeout:       durationEnv("READ_TIMEOUT", 15*time.Second),
		WriteTimeout:      durationEnv("WRITE_TIMEOUT", 20*time.Second),
		IdleTimeout:       durationEnv("IDLE_TIMEOUT", 60*time.Second),
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) ListenAddress() string { return ":" + c.Port }
func (c Config) IsProduction() bool    { return c.Environment == "production" }

func (c Config) RedisEnabled() bool { return strings.TrimSpace(c.RedisAddr) != "" }

func (c Config) Public() PublicConfig {
	return PublicConfig{
		AppName: c.AppName, Environment: c.Environment, DatabaseDriver: c.DatabaseDriver,
		RedisEnabled: c.RedisEnabled(), TokenTTLSeconds: int64(c.TokenTTL.Seconds()),
		RequestLimit: c.RequestLimit, ShutdownSeconds: int64(c.ShutdownTimeout.Seconds()),
	}
}

func (c Config) Validate() error {
	switch c.DatabaseDriver {
	case "postgres", "mysql", "sqlite":
	default:
		return fmt.Errorf("unsupported DATABASE_DRIVER %q", c.DatabaseDriver)
	}
	if strings.TrimSpace(c.AppName) == "" {
		return fmt.Errorf("APP_NAME must not be empty")
	}
	if strings.TrimSpace(c.Port) == "" {
		return fmt.Errorf("PORT must not be empty")
	}
	if len(c.JWTSecret) < 16 {
		return fmt.Errorf("JWT_SECRET must contain at least 16 characters")
	}
	if strings.Contains(c.MinIOEndpoint, "://") || c.MinIOEndpoint == "" || c.MinIOAccessKey == "" || c.MinIOSecretKey == "" || c.MinIOBucket == "" {
		return fmt.Errorf("MinIO endpoint must omit the URL scheme and all MinIO settings are required")
	}
	if c.RequestLimit < 10 || c.RequestLimit > 100000 {
		return fmt.Errorf("REQUEST_LIMIT must be between 10 and 100000")
	}
	for name, value := range map[string]time.Duration{
		"STARTUP_TIMEOUT": c.StartupTimeout, "SHUTDOWN_TIMEOUT": c.ShutdownTimeout,
		"READ_HEADER_TIMEOUT": c.ReadHeaderTimeout, "READ_TIMEOUT": c.ReadTimeout,
		"WRITE_TIMEOUT": c.WriteTimeout, "IDLE_TIMEOUT": c.IdleTimeout,
	} {
		if value <= 0 {
			return fmt.Errorf("%s must be positive", name)
		}
	}
	return nil
}

func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}

func intEnv(key string, fallback int) int {
	raw := env(key, "")
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func boolEnv(key string, fallback bool) bool {
	value, err := strconv.ParseBool(env(key, strconv.FormatBool(fallback)))
	if err != nil {
		return fallback
	}
	return value
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	raw := env(key, "")
	if raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	values := make([]string, 0)
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			values = append(values, item)
		}
	}
	return values
}
