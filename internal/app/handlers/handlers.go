package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/madatsci/urlshortener/internal/app/config"
	"github.com/madatsci/urlshortener/internal/app/models"
	"github.com/madatsci/urlshortener/internal/app/storage"
	"go.uber.org/zap"
)

type (
	Handlers struct {
		s   storage.Storage
		c   *config.Config
		log *zap.SugaredLogger
	}
)

// New creates new Handlers.
func New(config *config.Config, logger *zap.SugaredLogger, storage storage.Storage) (*Handlers, error) {
	return &Handlers{c: config, s: storage, log: logger}, nil
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

	shortURL, err := h.storeShortURL(r.Context(), url)
	if err != nil {
		h.handleError("AddHandler", err)
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
		h.handleError("AddHandlerJSON", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if request.URL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL, err := h.storeShortURL(r.Context(), request.URL)
	if err != nil {
		h.handleError("AddHandlerJSON", err)
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

	url, err := h.s.Get(r.Context(), slug)
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

func (h *Handlers) storeShortURL(ctx context.Context, longURL string) (string, error) {
	slug := generateSlug(slugLength)
	shortURL := fmt.Sprintf("%s/%s", h.c.BaseURL, slug)

	return shortURL, h.s.Add(ctx, slug, longURL)
}

func (h *Handlers) handleError(method string, err error) {
	h.log.Errorln("error handling request", "method", method, "err", err)
}
