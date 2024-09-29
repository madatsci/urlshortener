package app

import (
	"context"
	"time"

	"github.com/madatsci/urlshortener/internal/app/config"
	"github.com/madatsci/urlshortener/internal/app/logger"
	"github.com/madatsci/urlshortener/internal/app/server"
	"github.com/madatsci/urlshortener/internal/app/store"
	"github.com/madatsci/urlshortener/internal/app/store/database"
	fs "github.com/madatsci/urlshortener/internal/app/store/file"
	"github.com/madatsci/urlshortener/internal/app/store/memory"
	"go.uber.org/zap"
)

type (
	App struct {
		config *config.Config
		store  store.Store
		logger *zap.SugaredLogger
		server *server.Server
	}

	Options struct {
		ServerAddr      string
		BaseURL         string
		FileStoragePath string
		DatabaseDSN     string
		TokenSecret     []byte
		TokenDuration   time.Duration
	}
)

// New creates new App.
func New(ctx context.Context, opts Options) (*App, error) {
	config := config.New(opts.ServerAddr, opts.BaseURL, opts.FileStoragePath, opts.DatabaseDSN, opts.TokenSecret, opts.TokenDuration)

	logger, err := logger.New()
	if err != nil {
		return nil, err
	}

	store, err := newStore(ctx, config)
	if err != nil {
		return nil, err
	}

	srv := server.New(config, store, logger)

	app := &App{
		config: config,
		store:  store,
		logger: logger,
		server: srv,
	}

	return app, nil
}

// Start starts the application.
func (a *App) Start() error {
	return a.server.Start()
}

func newStore(ctx context.Context, config *config.Config) (store.Store, error) {
	if config.DatabaseDSN != "" {
		return database.New(ctx, config.DatabaseDSN)
	} else if config.FileStoragePath != "" {
		return fs.New(config.FileStoragePath)
	}

	return memory.New(), nil
}
