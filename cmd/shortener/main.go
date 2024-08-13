package main

import (
	"github.com/madatsci/urlshortener/internal/app/config"
	"github.com/madatsci/urlshortener/internal/app/server"
)

func main() {
	parseFlags()

	config := config.New(httpHost, httpPort, baseURL)

	s := server.New(config)
	if err := s.Start(); err != nil {
		panic(err)
	}
}
