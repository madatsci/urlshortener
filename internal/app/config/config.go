// Package config holds the service configuration.
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

const (
	tokenIssuer = "urlshortener"

	defaultServerAddr    = "localhost:8080"
	defaultBaseURL       = "http://localhost:8080"
	defaultTokenDuration = time.Hour
	defaultTokenSecret   = "secret_key"
)

// Config represents the service configuration.
type Config struct {
	ConfigFilePath string `envconfig:"CONFIG"`

	ServerAddr      string `envconfig:"SERVER_ADDRESS" json:"server_address"`
	EnableHTTPS     bool   `envconfig:"ENABLE_HTTPS" json:"enable_https"`
	BaseURL         string `envconfig:"BASE_URL" json:"base_url"`
	FileStoragePath string `envconfig:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DatabaseDSN     string `envconfig:"DATABASE_DSN" json:"database_dsn"`

	TokenSecret   string        `envconfig:"TOKEN_SECRET_KEY"`
	TokenDuration time.Duration `envconfig:"TOKEN_DURATION"`
	TokenIssuer   string

	enableHTTPSEnvSet  bool
	enableHTTPSFlagSet bool
}

func New() (*Config, error) {
	c, err := parseEnv()
	if err != nil {
		return nil, err
	}

	parseFlags(c)

	if err := parseFile(c); err != nil {
		return nil, err
	}

	setDefaults(c)

	return c, nil
}

func parseEnv() (*Config, error) {
	var c Config

	err := envconfig.Process("", &c)
	if err != nil {
		return nil, err
	}

	if _, set := os.LookupEnv("ENABLE_HTTPS"); set {
		c.enableHTTPSEnvSet = true
	}

	return &c, nil
}

func parseFlags(c *Config) {
	flag.Func("a", "address and port to run server in the form of host:port", func(flagValue string) error {
		if err := validateAddress(flagValue); err != nil {
			return fmt.Errorf("invalid server address: %s", err)
		}

		if c.ServerAddr == "" {
			c.ServerAddr = flagValue
		}
		return nil
	})

	flag.Func("b", "base URL of the generated short URL", func(flagValue string) error {
		u, err := url.Parse(flagValue)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return errors.New("invalid URL format")
		}

		if c.BaseURL == "" {
			c.BaseURL = flagValue
		}
		return nil
	})

	flag.Func("f", "file storage path", func(flagValue string) error {
		if flagValue == "" {
			return errors.New("invalid file path")
		}

		if c.FileStoragePath == "" {
			c.FileStoragePath = flagValue
		}
		return nil
	})

	flag.Func("d", "database DSN", func(flagValue string) error {
		if flagValue == "" {
			return errors.New("invalid database DSN")
		}

		if c.DatabaseDSN == "" {
			c.DatabaseDSN = flagValue
		}
		return nil
	})

	flag.Func("c", "config file path", func(flagValue string) error {
		if flagValue == "" {
			return errors.New("invalid file path")
		}

		if c.ConfigFilePath == "" {
			c.ConfigFilePath = flagValue
		}
		return nil
	})

	flag.Func("token-secret", "authentication token secret key", func(flagValue string) error {
		if flagValue == "" {
			return errors.New("invalid secret key")
		}

		if c.TokenSecret == "" {
			c.TokenSecret = flagValue
		}
		return nil
	})

	flag.Func("token-duration", "authentication token duration", func(flagValue string) error {
		if flagValue == "" {
			return errors.New("invalid duration")
		}

		duration, err := time.ParseDuration(flagValue)
		if err != nil {
			return errors.New("invalid duration")
		}

		if c.TokenDuration == 0 {
			c.TokenDuration = duration
		}
		return nil
	})

	flag.Func("s", "enable HTTPS", func(flagValue string) error {
		if !c.enableHTTPSEnvSet {
			var val bool
			if flagValue == "" {
				// -s with no value → true
				val = true
			} else {
				// -s=true, -s=1, etc. → parse it
				parsed, err := strconv.ParseBool(flagValue)
				if err != nil {
					return fmt.Errorf("invalid boolean value: %s", flagValue)
				}
				val = parsed
			}
			c.EnableHTTPS = val
			c.enableHTTPSFlagSet = true
		}
		return nil
	})

	flag.Parse()
}

func parseFile(c *Config) error {
	if c.ConfigFilePath == "" {
		return nil
	}

	var fileConfig Config

	data, err := os.ReadFile(c.ConfigFilePath)
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}

	if err := json.Unmarshal(data, &fileConfig); err != nil {
		return fmt.Errorf("unmarshal config from JSON: %w", err)
	}

	if c.ServerAddr == "" {
		c.ServerAddr = fileConfig.ServerAddr
	}
	if c.BaseURL == "" {
		c.BaseURL = fileConfig.BaseURL
	}
	if c.FileStoragePath == "" {
		c.FileStoragePath = fileConfig.FileStoragePath
	}
	if c.DatabaseDSN == "" {
		c.DatabaseDSN = fileConfig.DatabaseDSN
	}
	if !(c.enableHTTPSEnvSet || c.enableHTTPSFlagSet) {
		c.EnableHTTPS = fileConfig.EnableHTTPS
	}

	return nil
}

func setDefaults(c *Config) {
	if c.ServerAddr == "" {
		c.ServerAddr = defaultServerAddr
	}
	if c.BaseURL == "" {
		c.BaseURL = defaultBaseURL
	}
	if c.TokenDuration == 0 {
		c.TokenDuration = defaultTokenDuration
	}
	if c.TokenSecret == "" {
		c.TokenSecret = defaultTokenSecret
	}

	c.TokenIssuer = tokenIssuer
}

func validateAddress(value string) error {
	hp := strings.Split(value, ":")
	if len(hp) != 2 {
		return errors.New("wrong address format, must be host:port")
	}

	_, err := strconv.Atoi(hp[1])
	if err != nil {
		return errors.New("invalid port")
	}

	return nil
}
