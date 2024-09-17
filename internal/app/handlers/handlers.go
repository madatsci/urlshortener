package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/madatsci/urlshortener/internal/app/config"
	"github.com/madatsci/urlshortener/internal/app/models"
	"github.com/madatsci/urlshortener/internal/app/store"
	"go.uber.org/zap"
)

type (
	Handlers struct {
		s   store.Store
		c   *config.Config
		log *zap.SugaredLogger
	}
)

// New creates new Handlers.
func New(config *config.Config, logger *zap.SugaredLogger, store store.Store) (*Handlers, error) {
	return &Handlers{c: config, s: store, log: logger}, nil
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

		var alreadyExists *store.AlreadyExistsError
		if errors.As(err, &alreadyExists) {
			w.Header().Set("content-type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			shortURL = h.generateShortURLFromSlug(alreadyExists.URL.Short)
			w.Write([]byte(shortURL))
			return
		}
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

		var alreadyExists *store.AlreadyExistsError
		if errors.As(err, &alreadyExists) {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusConflict)

			response := models.ShortenResponse{
				Result: h.generateShortURLFromSlug(alreadyExists.URL.Short),
			}

			enc := json.NewEncoder(w)
			if err := enc.Encode(response); err != nil {
				panic(err)
			}
			return
		}
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

// TODO Add a test case for this.
func (h *Handlers) AddHandlerJSONBatch(w http.ResponseWriter, r *http.Request) {
	var request models.ShortenBatchRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&request); err != nil {
		h.handleError("AddHandlerJSONBatch", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(request.URLs) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	urls := make([]store.URL, 0, len(request.URLs))
	responseURLs := make([]models.ShortenBatchResponseItem, 0, len(request.URLs))
	for _, reqURL := range request.URLs {
		// TODO Fix copy-paste (see h.storeShortURL()).
		slug := generateSlug(slugLength)
		shortURL := h.generateShortURLFromSlug(slug)

		url := store.URL{
			ID:            uuid.NewString(),
			CorrelationID: reqURL.CorrelationID,
			Short:         slug,
			Original:      reqURL.OriginalURL,
			CreatedAt:     time.Now(),
		}
		urls = append(urls, url)

		responseURL := models.ShortenBatchResponseItem{
			CorrelationID: reqURL.CorrelationID,
			ShortURL:      shortURL,
		}
		responseURLs = append(responseURLs, responseURL)
	}

	err := h.s.AddBatch(r.Context(), urls)
	if err != nil {
		h.handleError("AddHandlerJSONBatch", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := &models.ShortenBatchRResponse{
		URLs: responseURLs,
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

	w.Header().Set("location", url.Original)
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

// Store returns Handlers' storage.
func (h *Handlers) Store() store.Store {
	return h.s
}

func (h *Handlers) storeShortURL(ctx context.Context, longURL string) (string, error) {
	slug := generateSlug(slugLength)
	shortURL := h.generateShortURLFromSlug(slug)

	url := store.URL{
		ID: uuid.NewString(),

		// TODO fix naming ambiguity:
		// slug is just a random string, while shortURL is the complete URL which contains the slug.
		Short:     slug,
		Original:  longURL,
		CreatedAt: time.Now(),
	}

	return shortURL, h.s.Add(ctx, url)
}

func (h *Handlers) generateShortURLFromSlug(slug string) string {
	return fmt.Sprintf("%s/%s", h.c.BaseURL, slug)
}

func (h *Handlers) handleError(method string, err error) {
	h.log.Errorln("error handling request", "method", method, "err", err)
}
