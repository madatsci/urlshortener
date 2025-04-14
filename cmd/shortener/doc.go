// Command cmd/shortener/shortener starts the URL shortener service.
//
// This service is a standalone web service which allows users to create short URLs
// that redirect to longer target URLs. It exposes a REST API for creating and
// resolving short links. Some API endpoints are protected with authorization.
//
// All data can be stored in memory (default), file, or database, depending on how
// the service is configured. The service can be configured via flags and environment
// variables. In case of conflict the value of the environment variable prevails.
//
// Example usage:
//
//	go run ./cmd/shortener/*.go
//
// Environment variables:
//
//	SERVER_ADDRESS    – Address and port to run server in the form of host:port (default: localhost:8080)
//	BASE_URL          - Base URL of the generated short URL
//	DATABASE_DSN      - Database DSN (in case you want to store data in database)
//	FILE_STORAGE_PATH - File storage path (in case you want to store data on disk)
//	TOKEN_SECRET_KEY  - Authentication token secret key
//	TOKEN_DURATION    - Authentication token duration (in the format of Golang duration string)
//
// Example:
//
//	curl -X POST http://localhost:8080/api/shorten -H "Content-Type: application/json" -d '{"url":"https://example.org"}'
//	=> {"result":"http://localhost:8080/bnwMHuSR"}
package main
