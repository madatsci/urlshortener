package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

	baseURL := "http://localhost"
	addr := ":8080"
	s := New(baseURL, addr)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.requestBody))
			w := httptest.NewRecorder()

			s.AddHandler(w, r)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode, "Unexpected response code")
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"), "Unexpected content type")

			defer res.Body.Close()

			if !test.want.wantErr {
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				var slug string
				for k := range s.s.ListAll() {
					slug = k
					break
				}

				wantBody := baseURL + addr + "/" + slug
				assert.Equal(t, wantBody, string(resBody))
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
				code:     http.StatusBadRequest,
				location: "",
				wantErr:  true,
			},
		},
	}

	baseURL := "http://localhost"
	addr := ":8080"
	s := New(baseURL, addr)
	s.s.Add("shortURL", "https://practicum.yandex.ru/")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, test.path, nil)
			w := httptest.NewRecorder()

			s.GetHandler(w, r)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode, "Unexpected response code")

			defer res.Body.Close()

			if !test.want.wantErr {
				assert.Equal(t, test.want.location, res.Header.Get("Location"), "Unexpected location")
			}
		})
	}
}
