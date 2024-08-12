package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/madatsci/urlshortener/internal/app/config"
	"github.com/madatsci/urlshortener/internal/app/handlers"
)

type Server struct {
	mux    *chi.Mux
	config *config.Config
	h      *handlers.Handlers
}

// New creates a new HTTP server.
func New(config *config.Config) *Server {
	server := &Server{config: config}

	r := chi.NewRouter()
	h := handlers.New(config)

	r.Route("/", func(r chi.Router) {
		r.Post("/", h.AddHandler)
		r.Post("/api/shorten", h.AddHandlerJSON)
		r.Get("/{slug}", h.GetHandler)
	})

	server.h = h
	server.mux = r

	return server
}

// Start starts the server under the specified address.
func (s *Server) Start() error {
	return http.ListenAndServe(fmt.Sprintf("%s:%d", s.config.HttpHost, s.config.HttpPort), s.mux)
}

// Router returns server router for usage in tests.
func (s *Server) Router() *chi.Mux {
	return s.mux
}
