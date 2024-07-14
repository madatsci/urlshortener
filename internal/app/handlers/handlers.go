package handlers

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/madatsci/urlshortener/internal/app/storage"
)

type Handlers struct {
	s       *storage.Storage
	baseURL string
	addr    string
}

// New creates new Handlers.
func New(baseURL, addr string) *Handlers {
	return &Handlers{baseURL: baseURL, addr: addr, s: storage.New()}
}

// AddHandler handles adding a new URL.
func (h *Handlers) AddHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	url := string(body)
	if url == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	slug := generateSlug(slugLength)
	shortURL := h.baseURL + h.addr + "/" + slug

	h.s.Add(slug, url)

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

// GetHandler handles retrieving the URL by its slug.
func (h *Handlers) GetHandler(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	url, err := h.s.Get(slug)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// Storage returns Handlers' storage.
func (h *Handlers) Storage() *storage.Storage {
	return h.s
}
