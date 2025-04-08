// Package server implements an HTTP web server with router and middleware.
package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/madatsci/urlshortener/internal/app/config"
	"github.com/madatsci/urlshortener/internal/app/handlers"
	mw "github.com/madatsci/urlshortener/internal/app/server/middleware"
	"github.com/madatsci/urlshortener/internal/app/store"
	"github.com/madatsci/urlshortener/pkg/jwt"
	"go.uber.org/zap"
)

// Server is the HTTP server for the URL shortener service.
//
// It holds the application's HTTP router, configuration, handler logic, and logger.
// Use New to create and configure a server instance, then call Start to begin
// handling HTTP requests.
type Server struct {
	mux    http.Handler
	config *config.Config
	h      *handlers.Handlers
	log    *zap.SugaredLogger
}

// New creates a new HTTP server.
func New(config *config.Config, store store.Store, logger *zap.SugaredLogger) *Server {
	server := &Server{
		config: config,
		log:    logger,
	}

	h := handlers.New(config, logger, store)

	r := chi.NewRouter()

	loggerMiddleware := mw.NewLogger(server.log)
	r.Use(loggerMiddleware.Logger)
	r.Use(mw.Gzip)
	r.Use(middleware.Recoverer)

	// Mounting net/http/pprof.
	r.Mount("/debug", middleware.Profiler())

	authMiddleware := mw.NewAuth(mw.Options{
		JWT: jwt.New(jwt.Options{
			Secret:   config.TokenSecret,
			Duration: config.TokenDuration,
			Issuer:   config.TokenIssuer,
		}),
		Store: store,
		Log:   logger,
	})

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.PublicAPIAuth)
		r.Post("/", h.AddHandler)
		r.Post("/api/shorten", h.AddHandlerJSON)
		r.Post("/api/shorten/batch", h.AddHandlerJSONBatch)
		// For some unknown reason Yandex Practicum tests now require
		// this endpoint to be public.
		// https://github.com/Yandex-Practicum/go-autotests/pull/82
		r.Get("/api/user/urls", h.GetUserURLsHandler)
	})

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.PrivateAPIAuth)
		r.Delete("/api/user/urls", h.DeleteUserURLsHandler)
	})

	r.Get("/ping", h.PingHandler)
	r.Get("/{slug}", h.GetHandler)

	server.h = h
	server.mux = r

	return server
}

// Start starts the server after it was created and configured.
func (s *Server) Start() error {
	s.log.Infof("starting server with config: %+v", s.config)

	return http.ListenAndServe(s.config.ServerAddr, s.mux)
}

// Router returns server router for usage in tests.
func (s *Server) Router() http.Handler {
	return s.mux
}
