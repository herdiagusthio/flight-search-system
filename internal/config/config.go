package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Server   ServerConfig
	Timeouts TimeoutConfig
	Retry    RetryConfig
	Logging  LoggingConfig
	App      AppConfig
}

type ServerConfig struct {
	Port         int           `env:"PORT" envDefault:"8080"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT" envDefault:"5s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" envDefault:"5s"`
}

type TimeoutConfig struct {
	GlobalSearch time.Duration `env:"GLOBAL_SEARCH_TIMEOUT" envDefault:"5s"`
	Provider     time.Duration `env:"PROVIDER_TIMEOUT" envDefault:"2s"`
}

type RetryConfig struct {
	MaxAttempts  int           `env:"RETRY_MAX_ATTEMPTS" envDefault:"3"`
	InitialDelay time.Duration `env:"RETRY_INITIAL_DELAY" envDefault:"100ms"`
	MaxDelay     time.Duration `env:"RETRY_MAX_DELAY" envDefault:"2s"`
	Multiplier   float64       `env:"RETRY_MULTIPLIER" envDefault:"2.0"`
}

type LoggingConfig struct {
	Level  string `env:"LOG_LEVEL" envDefault:"info"`
	Format string `env:"LOG_FORMAT" envDefault:"json"`
}

type AppConfig struct {
	Env string `env:"ENV" envDefault:"development"`
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Debug().Msg("No .env file found, using default environment values")
	}

	config := &Config{}
	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if err := validate(config); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return config, nil
}

func MustLoadConfig() *Config {
	config, err := LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	return config
}

// validate checks if the config values are valid
func validate(cfg *Config) error {
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d, must be between 1 and 65535", cfg.Server.Port)
	}

	// Validate timeout are positive
	if cfg.Server.ReadTimeout <= 0 {
		return fmt.Errorf("invalid read timeout: %v, must be positive", cfg.Server.ReadTimeout)
	}
	if cfg.Server.WriteTimeout <= 0 {
		return fmt.Errorf("invalid write timeout: %v, must be positive", cfg.Server.WriteTimeout)
	}
	if cfg.Timeouts.GlobalSearch <= 0 {
		return fmt.Errorf("invalid global search timeout: %v, must be positive", cfg.Timeouts.GlobalSearch)
	}
	if cfg.Timeouts.Provider <= 0 {
		return fmt.Errorf("invalid provider timeout: %v, must be positive", cfg.Timeouts.Provider)
	}

	// Validate provider timeout is less than global timeout
	if cfg.Timeouts.Provider >= cfg.Timeouts.GlobalSearch {
		return fmt.Errorf("PROVIDER_TIMEOUT (%s) should be less than GLOBAL_SEARCH_TIMEOUT (%s)",
			cfg.Timeouts.Provider, cfg.Timeouts.GlobalSearch)
	}

	// Validate retry configuration
	if cfg.Retry.MaxAttempts < 1 {
		return fmt.Errorf("RETRY_MAX_ATTEMPTS must be at least 1; got %d", cfg.Retry.MaxAttempts)
	}
	if cfg.Retry.InitialDelay < 0 {
		return fmt.Errorf("RETRY_INITIAL_DELAY must be non-negative; got %v", cfg.Retry.InitialDelay)
	}
	if cfg.Retry.MaxDelay < 0 {
		return fmt.Errorf("RETRY_MAX_DELAY must be non-negative; got %v", cfg.Retry.MaxDelay)
	}
	if cfg.Retry.MaxDelay > 0 && cfg.Retry.InitialDelay > cfg.Retry.MaxDelay {
		return fmt.Errorf("RETRY_INITIAL_DELAY (%v) should not exceed RETRY_MAX_DELAY (%v)",
			cfg.Retry.InitialDelay, cfg.Retry.MaxDelay)
	}
	if cfg.Retry.Multiplier < 1.0 {
		return fmt.Errorf("RETRY_MULTIPLIER must be at least 1.0; got %f", cfg.Retry.Multiplier)
	}

	// Validate log level
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[cfg.Logging.Level] {
		return fmt.Errorf("LOG_LEVEL must be one of: debug, info, warn, error; got %q", cfg.Logging.Level)
	}

	// Validate log format
	validFormats := map[string]bool{"json": true, "console": true}
	if !validFormats[cfg.Logging.Format] {
		return fmt.Errorf("LOG_FORMAT must be one of: json, console; got %q", cfg.Logging.Format)
	}

	// Validate app environment
	validEnvs := map[string]bool{"development": true, "staging": true, "production": true}
	if !validEnvs[cfg.App.Env] {
		return fmt.Errorf("APP_ENV must be one of: development, staging, production; got %q", cfg.App.Env)
	}

	return nil
}

