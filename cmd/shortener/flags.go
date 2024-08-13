package main

import (
	"errors"
	"flag"
	"regexp"
	"strconv"
	"strings"
)

var (
	httpHost = "localhost"
	httpPort = 8080
	baseURL  = "http://localhost:8080"
	validURL = regexp.MustCompile(`^https?:\/\/\w+:\d+`)
)

func parseFlags() {
	flag.Func("a", "address and port to run server in the form of host:port", func(flagValue string) error {
		return parseAddress(flagValue, &httpHost, &httpPort)
	})

	flag.Func("b", "base URL of the generated short URL", func(flagValue string) error {
		if !validURL.MatchString(flagValue) {
			return errors.New("invalid URL format")
		}

		baseURL = flagValue
		return nil
	})

	flag.Parse()
}

func parseAddress(value string, host *string, port *int) error {
	hp := strings.Split(value, ":")
	if len(hp) != 2 {
		return errors.New("wrong address format, must be host:port")
	}
	*host = hp[0]

	p, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}
	*port = p

	return nil
}
