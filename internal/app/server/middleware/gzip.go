package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// Gzip is a gzip encoding middleware handler.
func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		next.ServeHTTP(ow, r)
	})
}

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header is an implementation of http.ResponseWriter interface.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write is an implementation of http.ResponseWriter interface.
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader is an implementation of http.ResponseWriter interface.
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close closes the gzip Writer.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read implements io.Reader.
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close closes the Reader.
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
