package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/madatsci/urlshortener/internal/app/config"
	"github.com/madatsci/urlshortener/internal/app/models"
	"github.com/madatsci/urlshortener/internal/app/storage"
)

type (
	Handlers struct {
		s storage.Storage
		c *config.Config
	}
)

// New creates new Handlers.
func New(config *config.Config, storage storage.Storage) (*Handlers, error) {
	return &Handlers{c: config, s: storage}, nil
}

// AddHandler handles adding a new URL via text/plain request.
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

	shortURL, err := h.storeShortURL(url)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

// AddHandlerJSON handles adding a new URL via application/json request.
func (h *Handlers) AddHandlerJSON(w http.ResponseWriter, r *http.Request) {
	var request models.ShortenRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&request); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if request.URL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL, err := h.storeShortURL(request.URL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := models.ShortenResponse{
		Result: shortURL,
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(w)
	if err := enc.Encode(response); err != nil {
		panic(err)
	}
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

// PingHandler handles storage health-check.
func (h *Handlers) PingHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.s.Ping(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Storage returns Handlers' storage.
func (h *Handlers) Storage() storage.Storage {
	return h.s
}

func (h *Handlers) storeShortURL(longURL string) (string, error) {
	slug := generateSlug(slugLength)
	shortURL := fmt.Sprintf("%s/%s", h.c.BaseURL, slug)

	return shortURL, h.s.Add(slug, longURL)
}
