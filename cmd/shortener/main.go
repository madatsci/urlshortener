package main

import (
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const slugLength = 8

type apiHandler struct {
	urls map[string]string
}

func (a *apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		url := string(body)

		slug := generateRandomString(slugLength)
		shortURL := "http://localhost:8080/" + slug

		a.urls[slug] = url

		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	case http.MethodGet:
		shortURL := strings.Trim(r.URL.Path, "/")

		url, ok := a.urls[shortURL]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func main() {
	mux := http.NewServeMux()
	var handler apiHandler
	handler.urls = make(map[string]string)
	mux.Handle(`/`, &handler)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}

func generateRandomString(length int) string {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
