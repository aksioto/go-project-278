package config

import (
	"errors"
	"log"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

const (
	EnvProd = "prod"
	EnvDev  = "dev"
)

type Config struct {
	Env         string `env:"ENV" envDefault:"dev"`
	ServiceName string `env:"SERVICE_NAME,required"`

	Port    string `env:"PORT" envDefault:"80"`
	AppPort string `env:"APP_PORT" envDefault:"8080"`
	BaseURL string `env:"BASE_URL,required"`

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

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Env == "prod" && c.SentryDSN == "" {
		return errors.New("SENTRY_DSN is required in prod")
	}
	return nil
}

func (c *Config) IsProd() bool {
	return c.Env == EnvProd
}
