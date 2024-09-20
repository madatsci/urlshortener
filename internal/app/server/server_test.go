package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/madatsci/urlshortener/internal/app/config"
	"github.com/madatsci/urlshortener/internal/app/store"
	"github.com/madatsci/urlshortener/internal/app/store/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAddHandler(t *testing.T) {
	type want struct {
		code        int
		contentType string
		wantErr     bool
	}
	tests := []struct {
		name        string
		requestBody string
		want        want
	}{
		{
			name:        "positive case",
			requestBody: "https://practicum.yandex.ru/",
			want: want{
				code:        http.StatusCreated,
				contentType: "text/plain",
				wantErr:     false,
			},
		},
		{
			name:        "negative case: empty body",
			requestBody: "",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "",
				wantErr:     true,
			},
		},
	}

	s, ts := testServer(t)
	defer ts.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp := testRequest(t, ts, http.MethodPost, "/", strings.NewReader(test.requestBody))
			defer resp.Body.Close()

			assert.Equal(t, test.want.code, resp.StatusCode, "Unexpected response code")
			assert.Equal(t, test.want.contentType, resp.Header.Get("Content-Type"), "Unexpected content type")

			if !test.want.wantErr {
				respStr, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				wantBody := expectedShortURL(t, s, test.requestBody)
				assert.Equal(t, wantBody, string(respStr))
			}
		})
	}
}

func TestAddHandlerJSON(t *testing.T) {
	longURL := "https://practicum-yandex.ru"

	type want struct {
		code        int
		contentType string
		wantErr     bool
	}
	tests := []struct {
		name        string
		requestBody string
		want        want
	}{
		{
			name:        "positive case",
			requestBody: `{"url":"` + longURL + `"}`,
			want: want{
				code:        http.StatusCreated,
				contentType: "application/json",
				wantErr:     false,
			},
		},
		{
			name:        "negative case: invalid JSON",
			requestBody: "{",
			want: want{
				code:        http.StatusInternalServerError,
				contentType: "",
				wantErr:     true,
			},
		},
		{
			name:        "negative case: empty URL",
			requestBody: `{"url":""}`,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "",
				wantErr:     true,
			},
		},
	}

	s, ts := testServer(t)
	defer ts.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp := testRequest(t, ts, http.MethodPost, "/api/shorten", strings.NewReader(test.requestBody))
			defer resp.Body.Close()

			assert.Equal(t, test.want.code, resp.StatusCode, "Unexpected response code")
			assert.Equal(t, test.want.contentType, resp.Header.Get("Content-Type"), "Unexpected content type")

			if !test.want.wantErr {
				respStr, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				wantShortURL := expectedShortURL(t, s, longURL)
				wantBody := `{"result":"` + wantShortURL + `"}`
				assert.JSONEq(t, wantBody, string(respStr))
			}
		})
	}
}

func TestGetHandler(t *testing.T) {
	type want struct {
		code     int
		location string
		wantErr  bool
	}

	tests := []struct {
		name string
		path string
		want want
	}{
		{
			name: "positive case",
			path: "/shortURL",
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: "https://practicum.yandex.ru/",
				wantErr:  false,
			},
		},
		{
			name: "negative case: not found",
			path: "/wrongURL",
			want: want{
				code:     http.StatusBadRequest,
				location: "",
				wantErr:  true,
			},
		},
		{
			name: "negative case: empty path",
			path: "/",
			want: want{
				code:     http.StatusMethodNotAllowed,
				location: "",
				wantErr:  true,
			},
		},
	}

	s, ts := testServer(t)
	defer ts.Close()
	ctx := context.Background()

	longURL := "https://practicum.yandex.ru/"
	url := store.URL{
		ID:        uuid.NewString(),
		Short:     "shortURL",
		Original:  longURL,
		CreatedAt: time.Now(),
	}
	s.h.Store().Add(ctx, url)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp := testRequest(t, ts, http.MethodGet, test.path, nil)
			defer resp.Body.Close()

			assert.Equal(t, test.want.code, resp.StatusCode, "Unexpected response code")

			if !test.want.wantErr {
				assert.Equal(t, test.want.location, resp.Header.Get("Location"), "Unexpected location")
			}
		})
	}
}

func TestGzipCompression(t *testing.T) {
	s, ts := testServer(t)
	defer ts.Close()

	t.Run("sends_gzip", func(t *testing.T) {
		longURL := "https://practicum.yandex.ru/"
		requestBody := `{"url":"` + longURL + `"}`

		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/shorten", buf)
		require.NoError(t, err)
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "")

		resp := sendRequest(t, req)
		defer resp.Body.Close()

		respStr, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		wantShortURL := expectedShortURL(t, s, longURL)
		wantBody := `{"result":"` + wantShortURL + `"}`
		assert.JSONEq(t, wantBody, string(respStr))
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		longURL := "http://example.org"
		requestBody := `{"url":"` + longURL + `"}`

		buf := bytes.NewBufferString(requestBody)
		req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/shorten", buf)
		require.NoError(t, err)
		req.Header.Set("Accept-Encoding", "gzip")

		resp := sendRequest(t, req)
		defer resp.Body.Close()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		decompressedBody, err := io.ReadAll(zr)
		require.NoError(t, err)

		wantShortURL := expectedShortURL(t, s, longURL)
		wantBody := `{"result":"` + wantShortURL + `"}`
		assert.JSONEq(t, wantBody, string(decompressedBody))
	})
}

func testServer(t *testing.T) (*Server, *httptest.Server) {
	filepath := "../../../tmp/test_storage.txt"
	os.Remove(filepath)

	config := &config.Config{
		ServerAddr:      "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: filepath,
	}

	logger := zap.NewNop().Sugar()

	store, err := memory.New()
	if err != nil {
		panic(err)
	}

	s, err := New(config, store, logger)
	require.NoError(t, err)

	return s, httptest.NewServer(s.Router())
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) *http.Response {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)
	req.Header.Set("Accept-Encoding", "")

	return sendRequest(t, req)
}

func sendRequest(t *testing.T, req *http.Request) *http.Response {
	cli := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := cli.Do(req)
	require.NoError(t, err)

	return resp
}

func expectedShortURL(t *testing.T, s *Server, url string) string {
	var slug string
	for k, u := range s.h.Store().ListAll(context.Background()) {
		if u.Original == url {
			slug = k
			break
		}
	}

	if slug == "" {
		t.Errorf("url %s was not saved", url)
	}

	return fmt.Sprintf("%s/%s", s.config.BaseURL, slug)
}
