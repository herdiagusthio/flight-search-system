package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with all defaults",
			cfg: &Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     2 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				App: AppConfig{
					Env: "development",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid port - zero",
			cfg: &Config{
				Server: ServerConfig{
					Port:         0,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     2 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				App: AppConfig{
					Env: "development",
				},
			},
			wantErr: true,
			errMsg:  "invalid port: 0, must be between 1 and 65535",
		},
		{
			name: "invalid port - negative",
			cfg: &Config{
				Server: ServerConfig{
					Port:         -1,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     2 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				App: AppConfig{
					Env: "development",
				},
			},
			wantErr: true,
			errMsg:  "invalid port: -1, must be between 1 and 65535",
		},
		{
			name: "invalid port - too high",
			cfg: &Config{
				Server: ServerConfig{
					Port:         65536,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     2 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				App: AppConfig{
					Env: "development",
				},
			},
			wantErr: true,
			errMsg:  "invalid port: 65536, must be between 1 and 65535",
		},
		{
			name: "valid port - minimum",
			cfg: &Config{
				Server: ServerConfig{
					Port:         1,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     2 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				App: AppConfig{
					Env: "development",
				},
			},
			wantErr: false,
		},
		{
			name: "valid port - maximum",
			cfg: &Config{
				Server: ServerConfig{
					Port:         65535,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     2 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				App: AppConfig{
					Env: "development",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid read timeout - zero",
			cfg: &Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  0,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     2 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				App: AppConfig{
					Env: "development",
				},
			},
			wantErr: true,
			errMsg:  "invalid read timeout: 0s, must be positive",
		},
		{
			name: "invalid read timeout - negative",
			cfg: &Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  -1 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     2 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				App: AppConfig{
					Env: "development",
				},
			},
			wantErr: true,
			errMsg:  "invalid read timeout: -1s, must be positive",
		},
		{
			name: "invalid write timeout - zero",
			cfg: &Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 0,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     2 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				App: AppConfig{
					Env: "development",
				},
			},
			wantErr: true,
			errMsg:  "invalid write timeout: 0s, must be positive",
		},
		{
			name: "invalid global search timeout - zero",
			cfg: &Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 0,
					Provider:     2 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				App: AppConfig{
					Env: "development",
				},
			},
			wantErr: true,
			errMsg:  "invalid global search timeout: 0s, must be positive",
		},
		{
			name: "invalid provider timeout - zero",
			cfg: &Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     0,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				App: AppConfig{
					Env: "development",
				},
			},
			wantErr: true,
			errMsg:  "invalid provider timeout: 0s, must be positive",
		},
		{
			name: "invalid - provider timeout >= global search timeout",
			cfg: &Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     5 * time.Second, // equal to global
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				App: AppConfig{
					Env: "development",
				},
			},
			wantErr: true,
			errMsg:  "PROVIDER_TIMEOUT (5s) should be less than GLOBAL_SEARCH_TIMEOUT (5s)",
		},
		{
			name: "invalid - provider timeout > global search timeout",
			cfg: &Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     10 * time.Second, // greater than global
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				App: AppConfig{
					Env: "development",
				},
			},
			wantErr: true,
			errMsg:  "PROVIDER_TIMEOUT (10s) should be less than GLOBAL_SEARCH_TIMEOUT (5s)",
		},
		{
			name: "invalid log level",
			cfg: &Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     2 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "invalid",
					Format: "json",
				},
				App: AppConfig{
					Env: "development",
				},
			},
			wantErr: true,
			errMsg:  "LOG_LEVEL must be one of: debug, info, warn, error; got \"invalid\"",
		},
		{
			name: "valid log level - debug",
			cfg: &Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     2 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "debug",
					Format: "json",
				},
				App: AppConfig{
					Env: "development",
				},
			},
			wantErr: false,
		},
		{
			name: "valid log level - warn",
			cfg: &Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     2 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "warn",
					Format: "json",
				},
				App: AppConfig{
					Env: "development",
				},
			},
			wantErr: false,
		},
		{
			name: "valid log level - error",
			cfg: &Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     2 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "error",
					Format: "json",
				},
				App: AppConfig{
					Env: "development",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid log format",
			cfg: &Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     2 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "text", // invalid, should be json or console
				},
				App: AppConfig{
					Env: "development",
				},
			},
			wantErr: true,
			errMsg:  "LOG_FORMAT must be one of: json, console; got \"text\"",
		},
		{
			name: "valid log format - console",
			cfg: &Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     2 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "console",
				},
				App: AppConfig{
					Env: "development",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid app environment",
			cfg: &Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     2 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				App: AppConfig{
					Env: "test", // invalid
				},
			},
			wantErr: true,
			errMsg:  "APP_ENV must be one of: development, staging, production; got \"test\"",
		},
		{
			name: "valid app environment - staging",
			cfg: &Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     2 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				App: AppConfig{
					Env: "staging",
				},
			},
			wantErr: false,
		},
		{
			name: "valid app environment - production",
			cfg: &Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     2 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				App: AppConfig{
					Env: "production",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.cfg)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		wantCfg  *Config
		wantErr  bool
		errMatch string
	}{
		{
			name:    "default values when no env vars set",
			envVars: map[string]string{},
			wantCfg: &Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     2 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				App: AppConfig{
					Env: "development",
				},
			},
			wantErr: false,
		},
		{
			name: "custom port from env",
			envVars: map[string]string{
				"PORT": "3000",
			},
			wantCfg: &Config{
				Server: ServerConfig{
					Port:         3000,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     2 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				App: AppConfig{
					Env: "development",
				},
			},
			wantErr: false,
		},
		{
			name: "custom timeouts from env",
			envVars: map[string]string{
				"READ_TIMEOUT":          "10s",
				"WRITE_TIMEOUT":         "15s",
				"GLOBAL_SEARCH_TIMEOUT": "30s",
				"PROVIDER_TIMEOUT":      "5s",
			},
			wantCfg: &Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  10 * time.Second,
					WriteTimeout: 15 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 30 * time.Second,
					Provider:     5 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				App: AppConfig{
					Env: "development",
				},
			},
			wantErr: false,
		},
		{
			name: "custom logging config from env",
			envVars: map[string]string{
				"LOG_LEVEL":  "debug",
				"LOG_FORMAT": "console",
			},
			wantCfg: &Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     2 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "debug",
					Format: "console",
				},
				App: AppConfig{
					Env: "development",
				},
			},
			wantErr: false,
		},
		{
			name: "production environment",
			envVars: map[string]string{
				"ENV": "production",
			},
			wantCfg: &Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
				},
				Timeouts: TimeoutConfig{
					GlobalSearch: 5 * time.Second,
					Provider:     2 * time.Second,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				App: AppConfig{
					Env: "production",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid port - validation error",
			envVars: map[string]string{
				"PORT": "0",
			},
			wantErr:  true,
			errMatch: "validate config",
		},
		{
			name: "invalid port - too high",
			envVars: map[string]string{
				"PORT": "70000",
			},
			wantErr:  true,
			errMatch: "validate config",
		},
		{
			name: "invalid log level",
			envVars: map[string]string{
				"LOG_LEVEL": "invalid",
			},
			wantErr:  true,
			errMatch: "validate config",
		},
		{
			name: "invalid log format",
			envVars: map[string]string{
				"LOG_FORMAT": "xml",
			},
			wantErr:  true,
			errMatch: "validate config",
		},
		{
			name: "invalid app environment",
			envVars: map[string]string{
				"ENV": "test",
			},
			wantErr:  true,
			errMatch: "validate config",
		},
		{
			name: "provider timeout >= global search timeout",
			envVars: map[string]string{
				"GLOBAL_SEARCH_TIMEOUT": "2s",
				"PROVIDER_TIMEOUT":      "5s",
			},
			wantErr:  true,
			errMatch: "validate config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all relevant env vars before each test
			envVarsToClear := []string{
				"PORT", "READ_TIMEOUT", "WRITE_TIMEOUT",
				"GLOBAL_SEARCH_TIMEOUT", "PROVIDER_TIMEOUT",
				"LOG_LEVEL", "LOG_FORMAT", "ENV",
			}
			for _, key := range envVarsToClear {
				os.Unsetenv(key)
			}

			// Set test-specific env vars
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Cleanup after test
			t.Cleanup(func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			})

			cfg, err := LoadConfig()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMatch != "" {
					assert.Contains(t, err.Error(), tt.errMatch)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantCfg, cfg)
		})
	}
}
