// Package server implements an HTTP web server with router and middleware.
package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/madatsci/urlshortener/internal/app/config"
	"github.com/madatsci/urlshortener/internal/app/handlers"
	mw "github.com/madatsci/urlshortener/internal/app/server/middleware"
	"github.com/madatsci/urlshortener/internal/app/store"
	"github.com/madatsci/urlshortener/pkg/jwt"
)

// Server is the HTTP server for the URL shortener service.
//
// It holds the application's HTTP router, configuration, handler logic, and logger.
// Use New to create and configure a server instance, then call Start to begin
// handling HTTP requests.
type Server struct {
	mux        http.Handler
	config     *config.Config
	h          *handlers.Handlers
	log        *zap.SugaredLogger
	httpServer *http.Server
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
			Secret:   []byte(config.TokenSecret),
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
//
// It binds the listener and starts serving requests asynchronously. It returns
// nil once the listener is successfully bound; use Shutdown to gracefully stop
// the server.
func (s *Server) Start() error {
	s.log.Infof("starting server with config: %+v", s.config)

	server := &http.Server{
		Addr:    s.config.ServerAddr,
		Handler: s.mux,
	}
	s.httpServer = server

	ln, err := net.Listen("tcp", s.config.ServerAddr)
	if err != nil {
		return err
	}

	go func() {
		if s.config.EnableHTTPS {
			s.log.Info("HTTPS is enabled")
			cert, err := s.generateSelfSignedCert()
			if err != nil {
				s.log.Errorf("failed to generate self-signed certificate: %v", err)
				return
			}
			server.TLSConfig = &tls.Config{
				Certificates: []tls.Certificate{*cert},
			}

			tlsLn := tls.NewListener(ln, server.TLSConfig)
			if err := server.Serve(tlsLn); err != nil && !errors.Is(err, http.ErrServerClosed) {
				s.log.Errorf("HTTPS server error: %v", err)
			}
			return
		}

		if err := server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Errorf("server error: %v", err)
		}
	}()

	return nil
}

// Shutdown gracefully shuts down the server.
//
// It stops accepting new connections and waits for in-flight requests to
// complete, or until the provided context is done.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// Router returns server router for usage in tests.
func (s *Server) Router() http.Handler {
	return s.mux
}

// generateSelfSignedCert generates a self-signed TLS certificate for development purposes.
// Self-signed certificates aren't trusted by default because they're not issued
// by a recognized Certificate Authority (CA).
// To resolve this for testing with curl, you have a few options:
//  1. Use the --insecure (-k) flag with curl to bypass certificate validation.
//     For development purposes, it is the simplest and most common approach.
//  2. Modify the code to save the generated certificate to a file (e.g., server.crt) temporarily.
//     Then use `curl --cacert server.crt https://...`
//  3. Add the self-signed certificate to your system's trusted certificate store
//     (not recommended for general use).
func (s *Server) generateSelfSignedCert() (*tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Self-signed"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(time.Hour * 24 * 365),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:    []string{"localhost"},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	return &tlsCert, nil
}
