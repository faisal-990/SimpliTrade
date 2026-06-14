// Package config centralizes all environment-driven configuration into a single
// typed struct loaded once at startup. Nothing else in the codebase should read
// os.Getenv directly — depend on *Config instead, so settings are discoverable,
// testable, and validated in one place.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Env identifies the runtime environment. It governs safety defaults: in Prod we
// refuse to boot with insecure placeholders, in Dev we fill them with warnings.
type Env string

const (
	EnvDev  Env = "dev"
	EnvProd Env = "prod"
	EnvTest Env = "test"
)

// Config is the fully-resolved application configuration.
type Config struct {
	Env  Env
	HTTP HTTPConfig
	DB   DBConfig
	Auth AuthConfig
	// Market holds the external market-data provider settings. Keys are optional
	// today (FakeProvider needs none) and wired in at the end of the build.
	Market MarketConfig
	Engine EngineConfig
	Mail   MailConfig
	OAuth  OAuthConfig
}

// OAuthConfig holds external identity-provider credentials. Empty = that provider
// is simply not offered (the login button is hidden client-side).
type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
}

// MailConfig is the SMTP relay used for transactional email (e.g. password-reset
// codes). All empty = no email; the app falls back to logging codes in dev.
type MailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

type HTTPConfig struct {
	Port           string
	AllowedOrigins []string // CORS allowlist ("*" in dev)
	RateLimitRPS   int      // per-IP requests/second (0 = disabled)
	RateLimitBurst int      // per-IP burst allowance
	AppBaseURL     string   // public frontend URL, used to build links (e.g. password reset)
	APIBaseURL     string   // public URL of this server, used for OAuth redirect URIs
}

type DBConfig struct {
	Host     string
	User     string
	Password string
	Name     string
	Port     string
	SSLMode  string
	TimeZone string
}

// DSN renders the GORM/pgx connection string.
func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		d.Host, d.User, d.Password, d.Name, d.Port, d.SSLMode, d.TimeZone,
	)
}

type AuthConfig struct {
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type MarketConfig struct {
	Provider     string // "fake" until a real provider is wired in
	APIKey       string
	RatePerMin   int // max upstream calls/min (free tiers are tiny); 0 = unlimited
	RefreshLimit int // max symbols refreshed per engine cycle; 0 = all
	// Fundamentals can come from a different source than prices, since most free
	// price feeds gate fundamentals. Empty key = no separate source (price feed's
	// own fundamentals, if any, are used).
	FundamentalsProvider string // e.g. "finnhub"
	FundamentalsAPIKey   string
}

// EngineConfig governs the strategy daemon (Tower 2).
type EngineConfig struct {
	TickInterval  time.Duration // how often the engine refreshes + decides
	StrategiesDir string        // directory of investor strategy YAMLs
	// Sandbox swaps the real market refresher for a synthetic random-walk ticker,
	// so the daemon can run end-to-end without an external (rate-limited) feed —
	// useful for demos and when the real market is closed.
	Sandbox bool
	// IgnoreMarketHours runs the LIVE (real-provider) engine even when the market
	// is closed — for testing the real-data path off-hours. Production leaves it
	// false so the daemon trades only during the session.
	IgnoreMarketHours bool
}

// devJWTSecret is the only insecure fallback we tolerate, and only outside prod.
const devJWTSecret = "dev-insecure-jwt-secret-change-me"

// Load reads configuration from the process environment, optionally seeded by a
// .env file at the working directory (ignored if absent — production injects real
// env vars). It returns an error only for misconfigurations that are unsafe to
// run with (e.g. a missing JWT secret in prod), never merely for absent optional
// keys, so local/offline development needs no setup.
func Load() (*Config, error) {
	// Best-effort: a missing .env is normal (prod, CI, tests). Only surface real
	// parse errors.
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("config: loading .env: %w", err)
	}

	env := Env(strings.ToLower(getString("APP_ENV", string(EnvDev))))

	cfg := &Config{
		Env: env,
		HTTP: HTTPConfig{
			Port:           getString("PORT", "8080"),
			AllowedOrigins: getStringSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
			RateLimitRPS:   getInt("RATE_LIMIT_RPS", 20),
			RateLimitBurst: getInt("RATE_LIMIT_BURST", 40),
			AppBaseURL:     getString("APP_BASE_URL", "http://localhost:5173"),
			APIBaseURL:     getString("API_BASE_URL", "http://localhost:8080"),
		},
		DB: DBConfig{
			Host:     getString("HOST", "localhost"),
			User:     getString("DBUSER", "postgres"),
			Password: getString("DBPASS", "postgres"),
			Name:     getString("DBNAME", "simplitrade"),
			Port:     getString("DBPORT", "5432"),
			SSLMode:  getString("SSLMODE", "disable"),
			TimeZone: getString("TIMEZONE", "UTC"),
		},
		Auth: AuthConfig{
			JWTSecret:       getString("JWT_KEY", ""),
			AccessTokenTTL:  getDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
			RefreshTokenTTL: getDuration("REFRESH_TOKEN_TTL", 720*time.Hour), // 30 days
		},
		OAuth: OAuthConfig{
			GoogleClientID:     getString("GOOGLE_OAUTH_CLIENT_ID", ""),
			GoogleClientSecret: getString("GOOGLE_OAUTH_CLIENT_SECRET", ""),
		},
		Mail: MailConfig{
			Host:     getString("SMTP_HOST", ""),
			Port:     getString("SMTP_PORT", "587"),
			Username: getString("SMTP_USERNAME", ""),
			Password: getString("SMTP_PASSWORD", ""),
			From:     getString("SMTP_FROM", ""),
		},
		Market: MarketConfig{
			Provider:             getString("MARKET_PROVIDER", "fake"),
			APIKey:               getString("MARKET_API_KEY", ""),
			RatePerMin:           getInt("MARKET_RATE_PER_MIN", 8), // Twelve Data free tier
			RefreshLimit:         getInt("MARKET_REFRESH_LIMIT", 0), // 0 = all symbols
			FundamentalsProvider: getString("FUNDAMENTALS_PROVIDER", ""),
			FundamentalsAPIKey:   getString("FUNDAMENTALS_API_KEY", ""),
		},
		Engine: EngineConfig{
			TickInterval:      getDuration("ENGINE_TICK_INTERVAL", 60*time.Second),
			StrategiesDir:     getString("ENGINE_STRATEGIES_DIR", "internal/engine/strategies"),
			Sandbox:           getBool("ENGINE_SANDBOX", false),
			IgnoreMarketHours: getBool("ENGINE_IGNORE_MARKET_HOURS", false),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validate enforces invariants and fills safe dev fallbacks. It mutates cfg in
// place for the JWT secret fallback so callers always receive a usable value.
func (c *Config) validate() error {
	if c.Auth.JWTSecret == "" {
		if c.Env == EnvProd {
			return errors.New("config: JWT_KEY is required in prod")
		}
		// Dev/test convenience: deterministic secret so tokens work offline.
		c.Auth.JWTSecret = devJWTSecret
	}
	if c.Auth.AccessTokenTTL <= 0 {
		return errors.New("config: ACCESS_TOKEN_TTL must be positive")
	}
	if c.Auth.RefreshTokenTTL <= c.Auth.AccessTokenTTL {
		return errors.New("config: REFRESH_TOKEN_TTL must exceed ACCESS_TOKEN_TTL")
	}
	return nil
}

// IsProd reports whether the app is running in production.
func (c *Config) IsProd() bool { return c.Env == EnvProd }

func getString(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

// getStringSlice reads a comma-separated env var into a slice, trimming spaces.
func getStringSlice(key string, def []string) []string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return def
	}
	return out
}

func getInt(key string, def int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	if n, err := strconv.Atoi(v); err == nil {
		return n
	}
	return def
}

func getBool(key string, def bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	if b, err := strconv.ParseBool(v); err == nil {
		return b
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	// Accept Go duration strings ("15m", "720h") or a bare integer of seconds.
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	if secs, err := strconv.Atoi(v); err == nil {
		return time.Duration(secs) * time.Second
	}
	return def
}
