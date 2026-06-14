package config

import (
	"testing"
	"time"
)

// clearEnv unsets every key Load reads so each test starts from a known state,
// regardless of the developer's shell or a stray .env. t.Setenv restores values
// automatically at test end.
func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"APP_ENV", "PORT", "HOST", "DBUSER", "DBPASS", "DBNAME", "DBPORT",
		"SSLMODE", "TIMEZONE", "JWT_KEY", "ACCESS_TOKEN_TTL", "REFRESH_TOKEN_TTL",
		"MARKET_PROVIDER", "MARKET_API_KEY",
	} {
		t.Setenv(k, "")
	}
}

func TestLoad_DefaultsAreUsableOffline(t *testing.T) {
	clearEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() with no env should succeed in dev, got: %v", err)
	}
	if cfg.Env != EnvDev {
		t.Errorf("Env = %q, want dev", cfg.Env)
	}
	if cfg.HTTP.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.HTTP.Port)
	}
	if cfg.Market.Provider != "fake" {
		t.Errorf("Market.Provider = %q, want fake (no API key needed)", cfg.Market.Provider)
	}
	if cfg.Auth.JWTSecret == "" {
		t.Error("dev JWTSecret should fall back to a non-empty value")
	}
	if cfg.Auth.AccessTokenTTL != 15*time.Minute {
		t.Errorf("AccessTokenTTL = %v, want 15m", cfg.Auth.AccessTokenTTL)
	}
}

func TestLoad_ProdRequiresJWTSecret(t *testing.T) {
	clearEnv(t)
	t.Setenv("APP_ENV", "prod")

	if _, err := Load(); err == nil {
		t.Fatal("Load() in prod without JWT_KEY should error, got nil")
	}
}

func TestLoad_ProdWithSecretSucceeds(t *testing.T) {
	clearEnv(t)
	t.Setenv("APP_ENV", "prod")
	t.Setenv("JWT_KEY", "a-real-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() in prod with JWT_KEY should succeed, got: %v", err)
	}
	if !cfg.IsProd() {
		t.Error("IsProd() = false, want true")
	}
	if cfg.Auth.JWTSecret != "a-real-secret" {
		t.Errorf("JWTSecret = %q, want a-real-secret", cfg.Auth.JWTSecret)
	}
}

func TestLoad_DurationParsing(t *testing.T) {
	tests := []struct {
		name string
		val  string
		want time.Duration
	}{
		{"go duration", "30m", 30 * time.Minute},
		{"bare seconds", "90", 90 * time.Second},
		{"invalid falls back to default", "not-a-duration", 15 * time.Minute},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clearEnv(t)
			t.Setenv("ACCESS_TOKEN_TTL", tc.val)
			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load(): %v", err)
			}
			if cfg.Auth.AccessTokenTTL != tc.want {
				t.Errorf("AccessTokenTTL = %v, want %v", cfg.Auth.AccessTokenTTL, tc.want)
			}
		})
	}
}

func TestLoad_RefreshMustExceedAccess(t *testing.T) {
	clearEnv(t)
	t.Setenv("ACCESS_TOKEN_TTL", "10m")
	t.Setenv("REFRESH_TOKEN_TTL", "5m")

	if _, err := Load(); err == nil {
		t.Fatal("refresh TTL <= access TTL should error, got nil")
	}
}

func TestDSN_Format(t *testing.T) {
	d := DBConfig{
		Host: "h", User: "u", Password: "p", Name: "n",
		Port: "5432", SSLMode: "disable", TimeZone: "UTC",
	}
	want := "host=h user=u password=p dbname=n port=5432 sslmode=disable TimeZone=UTC"
	if got := d.DSN(); got != want {
		t.Errorf("DSN()\n got: %q\nwant: %q", got, want)
	}
}
