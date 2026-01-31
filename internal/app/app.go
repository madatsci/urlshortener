// Package app initializes and runs the URL shortener service.
//
// It wires together the configuration, storage layer, HTTP server,
// and logging components. The App type provides the entry point for
// starting the service. It still does not handle graceful shutdown yet.
package app

import (
	"context"

	"go.uber.org/zap"

	"github.com/madatsci/urlshortener/internal/app/config"
	"github.com/madatsci/urlshortener/internal/app/database"
	"github.com/madatsci/urlshortener/internal/app/logger"
	"github.com/madatsci/urlshortener/internal/app/server"
	"github.com/madatsci/urlshortener/internal/app/store"
	dbstore "github.com/madatsci/urlshortener/internal/app/store/database"
	fstore "github.com/madatsci/urlshortener/internal/app/store/file"
	memstore "github.com/madatsci/urlshortener/internal/app/store/memory"
)

// App is the top-level application container for the URL shortener service.
//
// Use New to create a new instance and Start to start the application.
type App struct {
	config *config.Config
	store  store.Store
	logger *zap.SugaredLogger
	server *server.Server

	buildVersion string
	buildDate    string
	buildCommit  string
}

// Options contains all dependencies required to build App.
type Options struct {
	Build  BuildOptions
	Config *config.Config
}

type BuildOptions struct {
	Version string
	Date    string
	Commit  string
}

// New creates a new App instance by initializing all core components,
// including the configuration, logger, storage layer, and HTTP server.
func New(ctx context.Context, opts Options) (*App, error) {
	logger, err := logger.New()
	if err != nil {
		return nil, err
	}

	store, err := newStore(ctx, opts.Config)
	if err != nil {
		return nil, err
	}

	srv := server.New(opts.Config, store, logger)

	app := &App{
		config:       opts.Config,
		store:        store,
		logger:       logger,
		server:       srv,
		buildVersion: opts.Build.Version,
		buildDate:    opts.Build.Date,
		buildCommit:  opts.Build.Commit,
	}

	return app, nil
}

// Start starts the URL shortener service and blocks until it is stopped.
func (a *App) Start() error {
	a.logger.Infof("Build version: %s", a.buildVersion)
	a.logger.Infof("Build date: %s", a.buildDate)
	a.logger.Infof("Build commit: %s", a.buildCommit)

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
