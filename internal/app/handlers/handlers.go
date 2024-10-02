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
	"github.com/madatsci/urlshortener/internal/app/server/middleware"
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
func New(config *config.Config, logger *zap.SugaredLogger, store store.Store) *Handlers {
	return &Handlers{c: config, s: store, log: logger}
}

// AddHandler handles adding a new URL via text/plain request.
func (h *Handlers) AddHandler(w http.ResponseWriter, r *http.Request) {
	userID := parseUserID(r)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	url := string(body)
	if url == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL, err := h.storeShortURL(r.Context(), url, userID)
	if err != nil {
		h.handleError("AddHandler", err)

		var alreadyExists *store.AlreadyExistsError
		if errors.As(err, &alreadyExists) {
			shortURL = h.generateShortURLFromSlug(alreadyExists.URL.Short)

			w.Header().Set("content-type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			if _, err := w.Write([]byte(shortURL)); err != nil {
				panic(err)
			}

			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.log.With("userID", userID).Info("new URL created")

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte(shortURL)); err != nil {
		panic(err)
	}
}

// AddHandlerJSON handles adding a new URL via application/json request.
func (h *Handlers) AddHandlerJSON(w http.ResponseWriter, r *http.Request) {
	userID := parseUserID(r)

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

	shortURL, err := h.storeShortURL(r.Context(), request.URL, userID)
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

	h.log.With("userID", userID).Info("new URL created")

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(w)
	if err := enc.Encode(response); err != nil {
		panic(err)
	}
}

// TODO Add a test case for this.
func (h *Handlers) AddHandlerJSONBatch(w http.ResponseWriter, r *http.Request) {
	userID := parseUserID(r)

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
			UserID:        userID,
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

	h.log.With("userID", userID, "count", len(urls)).Info("new URLs created via batch request")

	response := &models.ShortenBatchResponse{
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
		h.handleError("GetHandler", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if url.Deleted {
		w.WriteHeader(http.StatusGone)
		return
	}

	w.Header().Set("location", url.Original)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// GetUserURLsHandler handles retrieving all URLs created by the authorized user.
func (h *Handlers) GetUserURLsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := ensureUserID(r)
	if err != nil {
		h.log.With("handler", "GetUserURLsHandler").Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.log.With("userID", userID).Debug("fetching user urls")

	urls, err := h.s.ListByUserID(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	responseURLs := make([]models.UserURLItem, 0, len(urls))
	for _, url := range urls {
		responseURL := models.UserURLItem{
			ShortURL:    h.generateShortURLFromSlug(url.Short),
			OriginalURL: url.Original,
		}
		responseURLs = append(responseURLs, responseURL)
	}

	response := &models.ListByUserIDResponse{
		URLs: responseURLs,
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	if err := enc.Encode(response); err != nil {
		panic(err)
	}
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

func (h *Handlers) storeShortURL(ctx context.Context, longURL, userID string) (string, error) {
	slug := generateSlug(slugLength)
	shortURL := h.generateShortURLFromSlug(slug)

	url := store.URL{
		ID:     uuid.NewString(),
		UserID: userID,

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

func parseUserID(r *http.Request) string {
	userIDCtx := r.Context().Value(middleware.AuthenticatedUserKey)
	if userID, ok := userIDCtx.(string); ok {
		return userID
	}

	return ""
}

func ensureUserID(r *http.Request) (string, error) {
	userIDCtx := r.Context().Value(middleware.AuthenticatedUserKey)
	userID, ok := userIDCtx.(string)
	if !ok {
		return "", errors.New("authenticated user is required")
	}

	return userID, nil
}
