package main

import (
	"context"

	_ "net/http/pprof"

	"github.com/madatsci/urlshortener/internal/app"
)

func main() {
	if err := parseFlags(); err != nil {
		panic(err)
	}

	app, err := app.New(context.Background(), app.Options{
		ServerAddr:      serverAddr,
		BaseURL:         baseURL,
		FileStoragePath: fileStoragePath,
		DatabaseDSN:     databaseDSN,
		TokenSecret:     tokenSecret,
		TokenDuration:   tokenDuration,
	})
	if err != nil {
		panic(err)
	}

	if err = app.Start(); err != nil {
		panic(err)
	}
}
