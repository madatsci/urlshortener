package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/madatsci/urlshortener/internal/app/config"
	"github.com/madatsci/urlshortener/internal/app/handlers"
	"github.com/madatsci/urlshortener/internal/app/storage"
	"go.uber.org/zap"
)

type (
	Server struct {
		mux    *chi.Mux
		config *config.Config
		h      *handlers.Handlers
		log    *zap.SugaredLogger
	}
)

// New creates a new HTTP server.
func New(config *config.Config, logger *zap.SugaredLogger) (*Server, error) {
	server := &Server{
		config: config,
		log:    logger,
	}

	storage, err := storage.New(config)
	if err != nil {
		return nil, err
	}

	h, err := handlers.New(config, storage)
	if err != nil {
		return nil, err
	}

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", server.withMiddleware(h.AddHandler))
		r.Post("/api/shorten", server.withMiddleware(h.AddHandlerJSON))
		r.Get("/{slug}", server.withMiddleware(h.GetHandler))
	})

	server.h = h
	server.mux = r

	return server, nil
}

// Start starts the server under the specified address.
func (s *Server) Start() error {
	return http.ListenAndServe(s.config.ServerAddr, s.mux)
}

// Router returns server router for usage in tests.
func (s *Server) Router() *chi.Mux {
	return s.mux
}

func (s *Server) withMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return s.withLogging(gzipMiddleware(h))
}

func (s *Server) withLogging(h http.HandlerFunc) http.HandlerFunc {
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

func gzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		h.ServeHTTP(ow, r)
	}
}
