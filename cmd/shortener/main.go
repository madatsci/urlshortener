package main

import (
	"github.com/madatsci/urlshortener/internal/app/config"
	"github.com/madatsci/urlshortener/internal/app/logger"
	"github.com/madatsci/urlshortener/internal/app/server"
)

func main() {
	parseFlags()

	config := config.New(serverAddr, baseURL)

	logger, err := logger.New()
	if err != nil {
		panic(err)
	}

	s := server.New(config, logger)
	if err := s.Start(); err != nil {
		panic(err)
	}
}
