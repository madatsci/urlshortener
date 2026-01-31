package main

import (
	"context"

	_ "net/http/pprof"

	"github.com/madatsci/urlshortener/internal/app"
	"github.com/madatsci/urlshortener/internal/app/config"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	config, err := config.New()
	if err != nil {
		panic(err)
	}

	app, err := app.New(context.Background(), app.Options{
		Build: app.BuildOptions{
			Version: buildVersion,
			Date:    buildDate,
			Commit:  buildCommit,
		},
		Config: config,
	})
	if err != nil {
		panic(err)
	}

	if err = app.Start(); err != nil {
		panic(err)
	}
}
