package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/madatsci/urlshortener/internal/app/config"
	"github.com/madatsci/urlshortener/internal/app/storage"
)

type Handlers struct {
	s *storage.Storage
	c *config.Config
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Result string `json:"result"`
}

// New creates new Handlers.
func New(config *config.Config) *Handlers {
	return &Handlers{c: config, s: storage.New()}
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
	shortURL := fmt.Sprintf("%s/%s", h.c.BaseURL, slug)

	h.s.Add(slug, url)

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

// AddHandler handles adding a new URL via JSON request.
func (h *Handlers) AddHandlerJSON(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	var request shortenRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		panic(err)
	}

	if request.URL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	slug := generateSlug(slugLength)
	shortURL := fmt.Sprintf("%s/%s", h.c.BaseURL, slug)

	h.s.Add(slug, request.URL)

	response := &shortenResponse{
		Result: shortURL,
	}

	res, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(res)
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
