package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/madatsci/urlshortener/internal/app/handlers"
)

type Server struct {
	mux     *chi.Mux
	baseURL string
	addr    string
	h       *handlers.Handlers
}

// New creates a new HTTP server.
func New(baseURL, addr string) *Server {
	server := &Server{baseURL: baseURL, addr: addr}

	r := chi.NewRouter()
	h := handlers.New(baseURL, addr)

	r.Route("/", func(r chi.Router) {
		r.Post("/", h.AddHandler)
		r.Get("/{slug}", h.GetHandler)
	})

	server.h = h
	server.mux = r

	return server
}

// Start starts the server under the specified address.
func (s *Server) Start() error {
	return http.ListenAndServe(s.addr, s.mux)
}

// Router returns server router for usage in tests.
func (s *Server) Router() *chi.Mux {
	return s.mux
}
