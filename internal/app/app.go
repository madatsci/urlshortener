// Package app initializes and runs the URL shortener service.
//
// It wires together the configuration, storage layer, HTTP server,
// and logging components. The App type provides the entry point for
// starting the service and handles graceful shutdown on OS signals.
package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

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

// BuildOptions represents app build metadata.
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

// Start starts the URL shortener service and blocks until it is stopped by an
// OS signal.
//
// It binds the HTTP server and then waits for one of the handled signals
// (SIGINT, SIGTERM, SIGQUIT). On receiving a signal it gracefully shuts down
// the server, allowing in-flight requests to finish, and closes the storage,
// flushing any unsaved data.
func (a *App) Start() error {
	a.logger.Infof("Build version: %s", a.buildVersion)
	a.logger.Infof("Build date: %s", a.buildDate)
	a.logger.Infof("Build commit: %s", a.buildCommit)

	if err := a.server.Start(); err != nil {
		return err
	}
	a.logger.Info("server started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-quit

	a.logger.Infof("received signal %s, shutting down gracefully...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Errorf("server shutdown error: %v", err)
	}

	if err := a.store.Close(); err != nil {
		a.logger.Errorf("store close error: %v", err)
	}

	a.logger.Info("server stopped gracefully")
	return nil
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
