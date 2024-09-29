package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	serverAddr = "localhost:8080"
	baseURL    = "http://localhost:8080"

	tokenSecret   = []byte("secret_key")
	tokenDuration = time.Hour

	fileStoragePath, databaseDSN string
)

func parseFlags() error {
	flag.Func("a", "address and port to run server in the form of host:port", func(flagValue string) error {
		if err := validateAddress(flagValue); err != nil {
			return fmt.Errorf("invalid server address: %s", err)
		}

		serverAddr = flagValue
		return nil
	})

	flag.Func("b", "base URL of the generated short URL", func(flagValue string) error {
		u, err := url.Parse(flagValue)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return errors.New("invalid URL format")
		}

		baseURL = flagValue
		return nil
	})

	flag.Func("f", "file storage path", func(flagValue string) error {
		if flagValue == "" {
			return errors.New("invalid file path")
		}

		fileStoragePath = flagValue
		return nil
	})

	flag.Func("d", "database DSN", func(flagValue string) error {
		if flagValue == "" {
			return errors.New("invalid database DSN")
		}

		databaseDSN = flagValue
		return nil
	})

	flag.Func("token-secret", "authentication token secret key", func(flagValue string) error {
		if flagValue == "" {
			return errors.New("invalid secret key")
		}

		tokenSecret = []byte(flagValue)
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

		tokenDuration = duration
		return nil
	})

	flag.Parse()

	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		serverAddr = envServerAddress
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		baseURL = envBaseURL
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		fileStoragePath = envFileStoragePath
	}

	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		databaseDSN = envDatabaseDSN
	}

	if envTokenSecretKey := os.Getenv("TOKEN_SECRET_KEY"); envTokenSecretKey != "" {
		tokenSecret = []byte(envTokenSecretKey)
	}

	if envTokenDuration := os.Getenv("TOKEN_DURATION"); envTokenDuration != "" {
		duration, err := time.ParseDuration(envTokenDuration)
		if err != nil {
			return fmt.Errorf("invalid TOKEN_DURATION: %s", envTokenDuration)
		}

		tokenDuration = duration
	}

	return nil
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
