package app

import (
	"context"
	"time"

	"github.com/madatsci/urlshortener/internal/app/config"
	"github.com/madatsci/urlshortener/internal/app/database"
	"github.com/madatsci/urlshortener/internal/app/logger"
	"github.com/madatsci/urlshortener/internal/app/server"
	"github.com/madatsci/urlshortener/internal/app/store"
	dbstore "github.com/madatsci/urlshortener/internal/app/store/database"
	fstore "github.com/madatsci/urlshortener/internal/app/store/file"
	memstore "github.com/madatsci/urlshortener/internal/app/store/memory"
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
		conn, err := database.NewClient(ctx, config.DatabaseDSN)
		if err != nil {
			return nil, err
		}
		return dbstore.New(ctx, conn)
	} else if config.FileStoragePath != "" {
		return fstore.New(config.FileStoragePath)
	}

	return memstore.New(), nil
}
