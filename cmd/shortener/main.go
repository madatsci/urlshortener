package main

import (
	"context"

	"github.com/madatsci/urlshortener/internal/app/config"
	"github.com/madatsci/urlshortener/internal/app/logger"
	"github.com/madatsci/urlshortener/internal/app/server"
	"github.com/madatsci/urlshortener/internal/app/store"
	"github.com/madatsci/urlshortener/internal/app/store/database"
	fs "github.com/madatsci/urlshortener/internal/app/store/file"
	"github.com/madatsci/urlshortener/internal/app/store/memory"
)

func main() {
	parseFlags()

	config := config.New(serverAddr, baseURL, fileStoragePath, databaseDSN)

	logger, err := logger.New()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	store, err := newStore(ctx, config)
	if err != nil {
		panic(err)
	}

	// TODO Create app.go which incapsulates all deps (server, storage, logger, handlers, database client).
	s, err := server.New(ctx, config, store, logger)
	if err != nil {
		panic(err)
	}
	if err := s.Start(); err != nil {
		panic(err)
	}
}

// TODO Move this to somewhere (perhaps app.go).
func newStore(ctx context.Context, config *config.Config) (store.Store, error) {
	if config.DatabaseDSN != "" {
		return database.New(ctx, config.DatabaseDSN)
	} else if config.FileStoragePath != "" {
		return fs.New(config.FileStoragePath)
	}

	return memory.New()
}
