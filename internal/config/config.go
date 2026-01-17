package config

import (
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

const (
	EnvProd = "prod"
	EnvDev  = "dev"
)

type Config struct {
	Env         string `env:"ENV" envDefault:"dev"`
	ServiceName string `env:"SERVICE_NAME" envDefault:"Url Shortener"`
	AppVersion  string `env:"APP_VERSION" envDefault:"1.0.0"`

	Port       string `env:"PORT" envDefault:"80"`
	AppPort    string `env:"APP_PORT" envDefault:"8080"`
	RawBaseURL string `env:"BASE_URL,required"`
	BaseURL    *url.URL

	AllowedOrigins []string `env:"ALLOWED_ORIGINS" envDefault:"http://localhost:5173" envSeparator:","`

	DatabaseURL string `env:"DATABASE_URL,required"`
	MaxConns    int32  `env:"DATABASE_MAX_CONNS" envDefault:"10"`
	IdleTimeMs  int    `env:"DATABASE_IDLE_TIME_MS" envDefault:"30000"`

	SentryDSN            string  `env:"SENTRY_DSN"`
	SentrySampleRate     float64 `env:"SENTRY_SAMPLE_RATE" envDefault:"1.0"`
	SentryFlushTimeoutMs int     `env:"SENTRY_FLUSH_TIMEOUT_MS" envDefault:"2000"`
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	parsedURL, err := url.Parse(cfg.RawBaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid BASE_URL %q: %w", cfg.RawBaseURL, err)
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("BASE_URL must include scheme and host, got %q", cfg.RawBaseURL)
	}

	cfg.BaseURL = parsedURL

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Env == EnvProd && c.SentryDSN == "" {
		return errors.New("SENTRY_DSN is required in prod")
	}
	return nil
}

func (c *Config) IsProd() bool {
	return c.Env == EnvProd
}
