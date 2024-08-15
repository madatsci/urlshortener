package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/madatsci/urlshortener/internal/app/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	s, ts := testServer()
	defer ts.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, respStr := testRequest(t, ts, http.MethodPost, "/", strings.NewReader(test.requestBody))
			defer resp.Body.Close()

			assert.Equal(t, test.want.code, resp.StatusCode, "Unexpected response code")
			assert.Equal(t, test.want.contentType, resp.Header.Get("Content-Type"), "Unexpected content type")

			if !test.want.wantErr {
				var slug string
				for k := range s.h.Storage().ListAll() {
					slug = k
					break
				}

				wantBody := fmt.Sprintf("%s/%s", s.config.BaseURL, slug)
				assert.Equal(t, wantBody, respStr)
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

	s, ts := testServer()
	s.h.Storage().Add("shortURL", "https://practicum.yandex.ru/")
	defer ts.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, _ := testRequest(t, ts, http.MethodGet, test.path, nil)
			defer resp.Body.Close()

			assert.Equal(t, test.want.code, resp.StatusCode, "Unexpected response code")

			if !test.want.wantErr {
				assert.Equal(t, test.want.location, resp.Header.Get("Location"), "Unexpected location")
			}
		})
	}
}

func testServer() (*Server, *httptest.Server) {
	config := &config.Config{
		ServerAddr: "localhost:8080",
		BaseURL:    "http://localhost:8080",
	}

	s := New(config)

	return s, httptest.NewServer(s.Router())
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	cli := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := cli.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}
