package main

import (
	"context"

	_ "net/http/pprof"

	"github.com/madatsci/urlshortener/internal/app"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	if err := parseFlags(); err != nil {
		panic(err)
	}

	app, err := app.New(context.Background(), app.Options{
		BuildVersion:    buildVersion,
		BuildDate:       buildDate,
		BuildCommit:     buildCommit,
		ServerAddr:      serverAddr,
		BaseURL:         baseURL,
		FileStoragePath: fileStoragePath,
		DatabaseDSN:     databaseDSN,
		TokenSecret:     tokenSecret,
		TokenDuration:   tokenDuration,
		EnableHTTPS:     enableHTTPS,
	})
	if err != nil {
		panic(err)
	}

	if err = app.Start(); err != nil {
		panic(err)
	}
}
