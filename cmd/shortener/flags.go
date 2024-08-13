package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var (
	serverAddr = "localhost:8080"
	baseURL    = "http://localhost:8080"
)

func parseFlags() {
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

	flag.Parse()

	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		serverAddr = envServerAddress
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		baseURL = envBaseURL
	}
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
