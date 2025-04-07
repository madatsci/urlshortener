package server

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/madatsci/urlshortener/internal/app/config"
	"github.com/madatsci/urlshortener/internal/app/logger"
	"github.com/madatsci/urlshortener/internal/app/server/middleware"
	"github.com/madatsci/urlshortener/internal/app/store/memory"
)

func ExampleServer() {
	s, err := newExampleServer()
	if err != nil {
		log.Fatal(err)
	}

	// Add URL via plain text
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://example.org"))
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, r)
	fmt.Println(w.Code)

	// Created URL
	shortURL := w.Body.String()
	// Authorization token for future requests
	authToken := parseAuthToken(w.Result())
	w.Result().Body.Close()

	// Get full URL
	r = httptest.NewRequest(http.MethodGet, shortURL, nil)
	w = httptest.NewRecorder()
	s.mux.ServeHTTP(w, r)
	fmt.Println(w.Code)
	w.Result().Body.Close()

	// Add URL via JSON
	r = httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"http://example.org"}`))
	w = httptest.NewRecorder()
	s.mux.ServeHTTP(w, r)
	fmt.Println(w.Code)
	w.Result().Body.Close()

	// Add multiple URLs via JSON batch
	data := `
	[
        {"correlation_id":"mC9g8iasXW","original_url":"https://example.org/1"},
        {"correlation_id":"XFADu5Xlkw","original_url":"http://example.org/2"}
    ]
`
	r = httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(data))
	w = httptest.NewRecorder()
	s.mux.ServeHTTP(w, r)
	fmt.Println(w.Code)
	w.Result().Body.Close()

	// Get user URLs
	r = httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	r.AddCookie(&http.Cookie{
		Name:  middleware.DefaultCookieName,
		Value: authToken,
	})
	w = httptest.NewRecorder()
	s.mux.ServeHTTP(w, r)
	fmt.Println(w.Code)
	w.Result().Body.Close()

	// Delete user URLs
	r = httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(`["LduvFKkQ", "hVKwFYrF"]`))
	r.AddCookie(&http.Cookie{
		Name:  middleware.DefaultCookieName,
		Value: authToken,
	})
	w = httptest.NewRecorder()
	s.mux.ServeHTTP(w, r)
	fmt.Println(w.Code)
	w.Result().Body.Close()

	// Output:
	// 201
	// 307
	// 201
	// 201
	// 200
	// 202
}

func newExampleServer() (*Server, error) {
	c := &config.Config{
		BaseURL:       "http://localhost:8080",
		TokenDuration: time.Hour,
	}

	log, err := logger.New()
	if err != nil {
		return nil, err
	}

	return New(c, memory.New(), log), nil
}

func parseAuthToken(res *http.Response) string {
	for _, c := range res.Cookies() {
		if c.Name == middleware.DefaultCookieName {
			return c.Value
		}
	}

	return ""
}
