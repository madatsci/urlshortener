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

	authMiddleware := mw.NewAuth(mw.Options{
		JWT: jwt.New(jwt.Options{
			Secret:   config.TokenSecret,
			Duration: config.TokenDuration,
			Issuer:   config.TokenIssuer,
		}),
		Log: logger,
	})

	r.Route("/", func(r chi.Router) {
		// Public API
		r.Route("/", func(r chi.Router) {
			r.Use(authMiddleware.PublicAPIAuth)
			r.Post("/", h.AddHandler)
			r.Get("/{slug}", h.GetHandler)
			r.Post("/api/shorten", h.AddHandlerJSON)
			r.Post("/api/shorten/batch", h.AddHandlerJSONBatch)
		})

		// Private API
		r.Route("/api/user", func(r chi.Router) {
			r.Use(authMiddleware.PrivateAPIAuth)
			r.Get("/urls", h.GetUserURLsHandler)
			r.Delete("/urls", h.DeleteUserURLsHandler)
		})

		r.Get("/ping", h.PingHandler)
	})

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
