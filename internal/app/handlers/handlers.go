// Package handlers implements REST API request handlers.
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
	"go.uber.org/zap"

	"github.com/madatsci/urlshortener/internal/app/config"
	"github.com/madatsci/urlshortener/internal/app/models"
	"github.com/madatsci/urlshortener/internal/app/server/middleware"
	"github.com/madatsci/urlshortener/internal/app/store"
	"github.com/madatsci/urlshortener/pkg/random"
)

// Handlers is a service that provides HTTP handlers for REST API endpoints.
//
// It wires storage, configuration, and logger service. It also uses a channel
// for asynchronous processing requests for deleting URLs.
type Handlers struct {
	s   store.Store
	c   *config.Config
	log *zap.SugaredLogger

	delReqChan chan deleteURLRequest
}

type deleteURLRequest struct {
	userID string
	slug   string
}

const slugLength = 8

// New creates new Handlers.
func New(config *config.Config, logger *zap.SugaredLogger, store store.Store) *Handlers {
	h := &Handlers{
		c:          config,
		s:          store,
		log:        logger,
		delReqChan: make(chan deleteURLRequest, 1024),
	}

	go h.flushDeleteURLRequests(context.TODO())

	return h
}

// AddHandler handles adding a new URL via text/plain request.
func (h *Handlers) AddHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := ensureUserID(r)
	if err != nil {
		h.log.With("handler", "AddHandler").Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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
			shortURL = h.generateShortURLFromSlug(alreadyExists.URL.Slug)

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
	userID, err := ensureUserID(r)
	if err != nil {
		h.log.With("handler", "AddHandlerJSON").Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var request models.ShortenRequest
	dec := json.NewDecoder(r.Body)
	if err = dec.Decode(&request); err != nil {
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
				Result: h.generateShortURLFromSlug(alreadyExists.URL.Slug),
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

// AddHandlerJSONBatch handles adding a batch of URLs via application/json request.
func (h *Handlers) AddHandlerJSONBatch(w http.ResponseWriter, r *http.Request) {
	userID, err := ensureUserID(r)
	if err != nil {
		h.log.With("handler", "AddHandlerJSONBatch").Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var request models.ShortenBatchRequest
	dec := json.NewDecoder(r.Body)
	if err = dec.Decode(&request); err != nil {
		h.handleError("AddHandlerJSONBatch", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(request.URLs) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	urls := make([]models.URL, 0, len(request.URLs))
	responseURLs := make([]models.ShortenBatchResponseItem, 0, len(request.URLs))
	for _, reqURL := range request.URLs {
		if reqURL.OriginalURL == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// TODO Fix copy-paste (see h.storeShortURL()).
		slug := random.ASCIIString(slugLength)
		shortURL := h.generateShortURLFromSlug(slug)

		url := models.URL{
			ID:            uuid.NewString(),
			CorrelationID: reqURL.CorrelationID,
			Slug:          slug,
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

	err = h.s.BatchCreateURL(r.Context(), userID, urls)
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

	url, err := h.s.GetURL(r.Context(), slug)
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

	urls, err := h.s.ListURLsByUserID(r.Context(), userID)
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
			ShortURL:    h.generateShortURLFromSlug(url.Slug),
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

// DeleteUserURLsHandler deletes URLs with specified slugs created by the authorized user.
func (h *Handlers) DeleteUserURLsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := ensureUserID(r)
	if err != nil {
		h.log.With("handler", "DeleteUserURLsHandler").Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.log.With("userID", userID).Debug("deleting user urls")

	var request models.DeleteByUserIDRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&request); err != nil {
		h.handleError("DeleteUserURLsHandler", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(request.Slugs) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, slug := range request.Slugs {
		h.delReqChan <- deleteURLRequest{
			userID: userID,
			slug:   slug,
		}
	}

	w.WriteHeader(http.StatusAccepted)
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
	slug := random.ASCIIString(slugLength)
	shortURL := h.generateShortURLFromSlug(slug)

	url := models.URL{
		ID:        uuid.NewString(),
		Slug:      slug,
		Original:  longURL,
		CreatedAt: time.Now(),
	}

	return shortURL, h.s.CreateURL(ctx, userID, url)
}

func (h *Handlers) generateShortURLFromSlug(slug string) string {
	return fmt.Sprintf("%s/%s", h.c.BaseURL, slug)
}

func (h *Handlers) handleError(method string, err error) {
	h.log.Errorln("error handling request", "method", method, "err", err)
}

func (h *Handlers) flushDeleteURLRequests(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	var reqs []deleteURLRequest

	for {
		select {
		case req := <-h.delReqChan:
			reqs = append(reqs, req)
		case <-ticker.C:
			if len(reqs) == 0 {
				continue
			}

			for _, req := range reqs {
				if err := h.s.SoftDeleteURL(ctx, req.userID, req.slug); err != nil {
					h.log.Errorln("error deleting url", "err", err)
					continue
				}

				h.log.With("userID", req.userID, "slug", req.slug).Info("deleted url")
			}

			reqs = nil
		}
	}
}

func ensureUserID(r *http.Request) (string, error) {
	userIDCtx := r.Context().Value(middleware.AuthenticatedUserKey)
	userID, ok := userIDCtx.(string)
	if !ok {
		return "", errors.New("authenticated user is required")
	}

	return userID, nil
}
