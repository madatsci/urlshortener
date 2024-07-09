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
	mux     *http.ServeMux
	baseURL string
	addr    string
	s       *storage.Storage
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const slugLength = 8

// New creates a new HTTP server.
func New(baseURL, addr string) *Server {
	mux := http.NewServeMux()
	storage := storage.New()
	server := &Server{mux: mux, s: storage, baseURL: baseURL, addr: addr}

	mux.Handle("/", server.RootHandler())

	return server
}

// Start starts the server under the specified address.
func (s *Server) Start() error {
	return http.ListenAndServe(s.addr, s.mux)
}

// RootHandler handles basic request.
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

// AddHandler handles adding a new URL.
func (s *Server) AddHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	url := string(body)
	if url == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	slug := generateRandomString(slugLength)
	shortURL := s.baseURL + s.addr + "/" + slug

	s.s.Add(slug, url)

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

// GetHandler handles retrieving the URL by its slug.
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
