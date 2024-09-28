package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/madatsci/urlshortener/internal/app/config"
	"github.com/madatsci/urlshortener/internal/app/handlers"
	mw "github.com/madatsci/urlshortener/internal/app/server/middleware"
	"github.com/madatsci/urlshortener/internal/app/store"
	"go.uber.org/zap"
)

type (
	Server struct {
		mux    http.Handler
		config *config.Config
		h      *handlers.Handlers
		log    *zap.SugaredLogger
	}
)

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

	r.Route("/", func(r chi.Router) {
		r.Post("/", h.AddHandler)
		r.Post("/api/shorten", h.AddHandlerJSON)
		r.Post("/api/shorten/batch", h.AddHandlerJSONBatch)
		r.Get("/{slug}", h.GetHandler)
		r.Get("/ping", h.PingHandler)
	})

	// TODO EnsureAuth (not always)
	//create a new request context containing the authenticated user
	//ctxWithUser := context.WithValue(r.Context(), authenticatedUserKey, user)
	//create a new request using that new context
	//rWithUser := r.WithContext(ctxWithUser)

	server.h = h
	server.mux = r

	return server
}

// Start starts the server under the specified address.
func (s *Server) Start() error {
	s.log.Infof("starting server with config: %+v", s.config)

	return http.ListenAndServe(s.config.ServerAddr, s.mux)
}

// Router returns server router for usage in tests.
func (s *Server) Router() http.Handler {
	return s.mux
}
