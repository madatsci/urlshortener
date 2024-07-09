package main

import (
	"github.com/madatsci/urlshortener/internal/app/server"
)

func main() {
	s := server.New()
	if err := s.Start(":8080"); err != nil {
		panic(err)
	}
}
