// Package main is used for testing urlshortener API, especially for profiling.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/madatsci/urlshortener/internal/app/server/middleware"
	"github.com/madatsci/urlshortener/pkg/random"
)

type (
	createURLPlainTextResponse string

	createURLJSONResponse struct {
		Result string `json:"result"`
	}

	createURLJSONBatchResponse []createURLJSONBatchResponseItem

	createURLJSONBatchResponseItem struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}

	getUserURLsResponse []getUserURLsResponseItem

	getUserURLsResponseItem struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}
)

const endpoint = "http://localhost:8080"

func main() {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	ptRes, _ := createURLPlainText(client)
	getURL(client, string(ptRes))

	jRes, _ := createURLJSON(client, "")
	getURL(client, jRes.Result)

	jbRes, authToken := createURLJSONBatch(client, "")
	for _, item := range jbRes {
		getURL(client, item.ShortURL)
	}

	jRes, _ = createURLJSON(client, authToken)
	getURL(client, jRes.Result)
	getUserURLs(client, authToken)
	deleteUserURLs(client, authToken, []string{getSlugFromURL(jRes.Result)})
	getURL(client, jRes.Result)
	time.Sleep(15 * time.Second)
	getUserURLs(client, authToken)
	getURL(client, jRes.Result)

	jbRes, _ = createURLJSONBatch(client, authToken)
	getUserURLs(client, authToken)
	slugs := make([]string, 0, len(jbRes))
	for _, item := range jbRes {
		slugs = append(slugs, getSlugFromURL(item.ShortURL))
	}
	deleteUserURLs(client, authToken, slugs)
	time.Sleep(15 * time.Second)
	for _, item := range jbRes {
		getURL(client, item.ShortURL)
	}
	getUserURLs(client, authToken)
}

func createURLPlainText(client *http.Client) (createURLPlainTextResponse, string) {
	fmt.Println("\n====== Create URL via text/plain ======")
	urlString := random.URL().String()

	request := buildRequest(http.MethodPost, endpoint, strings.NewReader(urlString), "")
	request.Header.Add("Content-Type", "text/plain")

	return doTextRequest(client, request)
}

func createURLJSON(client *http.Client, authToken string) (createURLJSONResponse, string) {
	fmt.Println("\n====== Create URL via application/json ======")
	if authToken != "" {
		fmt.Println("Using auth token:", authToken)
	}

	urlString := random.URL().String()

	data := map[string]string{
		"url": urlString,
	}
	reqBody, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	request := buildRequest(http.MethodPost, endpoint+"/api/shorten", bytes.NewReader(reqBody), authToken)
	request.Header.Add("Content-Type", "application/json")

	var res createURLJSONResponse
	tokenFromHeader := doJSONRequest(client, request, &res)

	return res, tokenFromHeader
}

func createURLJSONBatch(client *http.Client, authToken string) (createURLJSONBatchResponse, string) {
	fmt.Println("\n====== Create URL batch via application/json ======")
	if authToken != "" {
		fmt.Println("Using auth token:", authToken)
	}

	type urlItem struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}

	batchSize := 20

	data := make([]urlItem, 0, batchSize)
	for i := 0; i < batchSize; i++ {
		data = append(data, urlItem{
			CorrelationID: random.ASCIIString(10),
			OriginalURL:   random.URL().String(),
		})
	}
	reqBody, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	request := buildRequest(http.MethodPost, endpoint+"/api/shorten/batch", bytes.NewReader(reqBody), authToken)
	request.Header.Add("Content-Type", "application/json")

	var res createURLJSONBatchResponse
	tokenFromHeader := doJSONRequest(client, request, &res)

	return res, tokenFromHeader
}

func getURL(client *http.Client, url string) {
	fmt.Println("\n====== Get URL ======")
	fmt.Println("URL:", url)

	request := buildRequest(http.MethodGet, url, nil, "")
	doTextRequest(client, request)
}

func getUserURLs(client *http.Client, authToken string) (getUserURLsResponse, string) {
	fmt.Println("\n====== Get user URLs ======")
	fmt.Printf("Using auth token: %s\n", authToken)

	request := buildRequest(http.MethodGet, endpoint+"/api/user/urls", nil, authToken)

	var res getUserURLsResponse
	doJSONRequest(client, request, &res)

	return res, ""
}

func deleteUserURLs(client *http.Client, authToken string, slugs []string) {
	fmt.Println("\n====== Delete user URL ======")
	fmt.Println("Slugs:", slugs)

	reqBody, err := json.Marshal(slugs)
	if err != nil {
		panic(err)
	}
	request := buildRequest(http.MethodDelete, endpoint+"/api/user/urls", bytes.NewReader(reqBody), authToken)
	request.Header.Add("Content-Type", "application/json")

	doJSONRequest(client, request, nil)
}

func buildRequest(method, url string, body io.Reader, authToken string) *http.Request {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(err)
	}
	if authToken != "" {
		request.AddCookie(&http.Cookie{
			Name:  "auth_token",
			Value: authToken,
		})
	}

	return request
}

func doTextRequest(client *http.Client, request *http.Request) (createURLPlainTextResponse, string) {
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nResponse code: %s\n", response.Status)
	defer response.Body.Close()

	var res createURLPlainTextResponse
	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println("\nResponse:")
	res = createURLPlainTextResponse(respBody)
	fmt.Println(res)
	fmt.Println()

	return res, parseAuthToken(response)
}

func doJSONRequest(client *http.Client, request *http.Request, result interface{}) string {
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nResponse code: %s\n", response.Status)
	defer response.Body.Close()

	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println("\nResponse:")
	fmt.Println(string(respBody))
	if response.StatusCode == 200 || response.StatusCode == 201 {
		err = json.Unmarshal(respBody, result)
		if err != nil {
			panic(err)
		}
	}

	return parseAuthToken(response)
}

func parseAuthToken(res *http.Response) string {
	for _, c := range res.Cookies() {
		if c.Name == middleware.DefaultCookieName {
			fmt.Printf("Auth token:\n%s\n", c.Value)

			return c.Value
		}
	}

	return ""
}

func getSlugFromURL(url string) string {
	return strings.Replace(url, endpoint+"/", "", 1)
}
