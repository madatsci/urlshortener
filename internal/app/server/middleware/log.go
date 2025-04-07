package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Logger is a logger middleware.
//
// Use NewLogger to create a new instance of Logger.
type Logger struct {
	log *zap.SugaredLogger
}

// NewLogger creates a new instance of Logger.
func NewLogger(log *zap.SugaredLogger) *Logger {
	return &Logger{log: log}
}

// Logger defines a Logger middleware handler.
func (l *Logger) Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		next.ServeHTTP(&lw, r)

		duration := time.Since(start)

		l.log.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	})
}

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Write writes data to the connection and calculates cumulative data size.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader sends an HTTP response header with the provided status code.
// It also saves the status code in the receiver.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
