package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App      AppConfig      `env-prefix:"APP_"`
	Database DatabaseConfig `env-prefix:"DB_"`
}

type AppConfig struct {
	Port int `env:"PORT" envDefault:"8080"`
}

type DatabaseConfig struct {
	Host         string        `env:"HOST" env-required:"true"`
	Port         int           `env:"PORT" env-required:"true"`
	User         string        `env:"USER" env-required:"true"`
	Password     string        `env:"PASSWORD" env-required:"true"`
	DBName       string        `env:"NAME" env-required:"true"`
	SSLMode      string        `env:"SSLMODE" env-required:"true"`
	MaxOpenConns int           `env:"MAX_OPEN_CONNS" env-default:"10"`
	MaxIdleConns int           `env:"MAX_IDLE_CONNS" env-default:"5"`
	MaxLifetime  time.Duration `env:"MAX_LIFETIME" env-default:"5m"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read .env: %w", err)
	}

	if err := cfg.App.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate app config: %w", err)
	}

	if err := cfg.Database.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate database config: %w", err)
	}

	return &cfg, nil
}

func (c *AppConfig) Validate() error {
	if c.Port <= 0 {
		return fmt.Errorf("app port must be greater than 0")
	}
	return nil
}

func (c *DatabaseConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	if c.Port <= 0 {
		return fmt.Errorf("port must be greater than 0")
	}

	if c.User == "" {
		return fmt.Errorf("user cannot be empty")
	}

	if c.Password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	if c.DBName == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	if c.SSLMode == "" {
		return fmt.Errorf("ssl mode cannot be empty")
	}

	if c.MaxOpenConns <= 0 {
		return fmt.Errorf("max open connections must be greater than 0")
	}

	if c.MaxIdleConns <= 0 {
		return fmt.Errorf("max idle connections must be greater than 0")
	}

	if c.MaxLifetime <= 0 {
		return fmt.Errorf("max lifetime must be greater than 0")
	}

	return nil
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}
