package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type gzipWriter struct {
	http.ResponseWriter
	writer io.Writer
}

func (w gzipWriter) Write(p []byte) (int, error) {
	return w.writer.Write(p)
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzipReader, err := gzip.NewReader(r.Body)
			if err != nil {
				zap.L().Error("Failed to read gzipped request body", zap.Error(err))
				http.Error(w, "Failed to read gzipped request body", http.StatusBadRequest)
				return
			}
			defer func(gzipReader *gzip.Reader) {
				err := gzipReader.Close()
				if err != nil {
					zap.L().Error("Failed to close gzipped request body", zap.Error(err))
					return
				}
			}(gzipReader)
			r.Body = gzipReader
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		contentType := r.Header.Get("Content-Type")
		if !isCompressibleContentType(contentType) {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			zap.L().Error("Failed to create gzip writer", zap.Error(err))
			http.Error(w, "Failed to create gzip writer", http.StatusInternalServerError)
			return
		}
		defer func(gz *gzip.Writer) {
			err := gz.Close()
			if err != nil {
				zap.L().Error("Failed to close gzip writer", zap.Error(err))
				return
			}
		}(gz)

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, writer: gz}, r)
	})
}

func isCompressibleContentType(contentType string) bool {
	compressibleTypes := []string{
		"application/json",
		"text/html",
	}

	for _, t := range compressibleTypes {
		if strings.Contains(contentType, t) {
			return true
		}
	}
	return false
}
