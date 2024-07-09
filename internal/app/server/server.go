package server

import (
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/madatsci/urlshortener/internal/app/storage"
)

type Server struct {
	mux *http.ServeMux
	s   *storage.Storage
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const slugLength = 8

// New creates a new HTTP server.
func New() *Server {
	mux := http.NewServeMux()
	storage := storage.New()
	server := &Server{mux: mux, s: storage}

	mux.Handle("/", server.RootHandler())

	return server
}

// Start starts the server under the specified address.
func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) RootHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			s.AddHandler(w, r)
		case http.MethodGet:
			s.GetHandler(w, r)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	})
}

func (s *Server) AddHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	url := string(body)

	slug := generateRandomString(slugLength)
	shortURL := "http://localhost:8080/" + slug

	s.s.Add(slug, url)

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (s *Server) GetHandler(w http.ResponseWriter, r *http.Request) {
	slug := strings.Trim(r.URL.Path, "/")

	url, err := s.s.Get(slug)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func generateRandomString(length int) string {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
