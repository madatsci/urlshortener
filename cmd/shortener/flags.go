package main

import (
	"errors"
	"flag"
	"strconv"
	"strings"
)

var httpHost, generatedHost string
var httpPort, generatedPort int

func parseFlags() {
	httpHost = "localhost"
	httpPort = 8080
	generatedHost = "localhost"
	generatedPort = 8080

	flag.Func("a", "address and port to run server in the form of host:port", func(flagValue string) error {
		return parseAddress(flagValue, &httpHost, &httpPort)
	})

	flag.Func("b", "base address of the generated short URL in the form of host:port", func(flagValue string) error {
		return parseAddress(flagValue, &generatedHost, &generatedPort)
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
