package main

import (
	"github.com/madatsci/urlshortener/internal/app/server"
)

func main() {
	s := server.New("http://localhost", ":8080")
	if err := s.Start(); err != nil {
		panic(err)
	}
}
