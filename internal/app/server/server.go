package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/madatsci/urlshortener/internal/app/config"
	"github.com/madatsci/urlshortener/internal/app/handlers"
	"go.uber.org/zap"
)

type (
	Server struct {
		mux    *chi.Mux
		config *config.Config
		h      *handlers.Handlers
		log    *zap.SugaredLogger
	}

	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// New creates a new HTTP server.
func New(config *config.Config, logger *zap.SugaredLogger) *Server {
	server := &Server{
		config: config,
		log:    logger,
	}

	r := chi.NewRouter()
	h := handlers.New(config)

	r.Route("/", func(r chi.Router) {
		r.Post("/", server.withLogging(h.AddHandler))
		r.Post("/api/shorten", h.AddHandlerJSON)
		r.Get("/{slug}", server.withLogging(h.GetHandler))
	})

	server.h = h
	server.mux = r

	return server
}

// Start starts the server under the specified address.
func (s *Server) Start() error {
	return http.ListenAndServe(s.config.ServerAddr, s.mux)
}

// Router returns server router for usage in tests.
func (s *Server) Router() *chi.Mux {
	return s.mux
}

func (s *Server) withLogging(f func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	h := http.HandlerFunc(f)

	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		s.log.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	}

	return logFn
}
