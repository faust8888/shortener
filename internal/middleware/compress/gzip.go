package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// compressWriter — это обёртка над http.ResponseWriter, которая сжимает данные с использованием gzip.
type compressWriter struct {
	w  http.ResponseWriter // Исходный ResponseWriter
	zw *gzip.Writer        // Gzip-писатель для сжатия данных
}

// newCompressWriter создаёт новый экземпляр compressWriter.
func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header возвращает заголовки исходного ResponseWriter.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write записывает данные в gzip.Writer, сжимая их перед отправкой клиенту.
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader устанавливает заголовок Content-Encoding: gzip и вызывает WriteHeader у исходного ResponseWriter.
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer, завершая поток сжатия.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// compressReader — это обёртка над io.ReadCloser, которая распаковывает gzip-сжатые данные из тела запроса.
type compressReader struct {
	r  io.ReadCloser // Исходный ReadCloser
	zr *gzip.Reader  // Gzip-читатель для распаковки данных
}

// newCompressReader создаёт новый экземпляр compressReader.
// Возвращает ошибку, если не удалось создать gzip.Reader.
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

// Read читает распакованные данные из gzip.Reader.
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close закрывает оба Reader'а: gzip.Reader и исходный ReadCloser.
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// NewMiddleware возвращает HTTP middleware, который:
// - Сжимает ответ сервера, если клиент поддерживает gzip.
// - Распаковывает входящие тела запроса, если они сжаты gzip.
func NewMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		r := res

		if isCompressUsingForResponse(res, req) {
			compressResponseWriter := newCompressWriter(res)
			r = compressResponseWriter
			defer compressResponseWriter.Close()
		}

		if isRequestSentWithEncoding(req) {
			compressRequestReader, err := newCompressReader(req.Body)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
			req.Body = compressRequestReader
			defer compressRequestReader.Close()
		}
		handler.ServeHTTP(r, req)
	})
}

// isRequestSentWithEncoding проверяет, было ли тело запроса отправлено с использованием gzip-сжатия.
func isRequestSentWithEncoding(r *http.Request) bool {
	contentEncoding := r.Header.Get("Content-Encoding")
	return strings.Contains(contentEncoding, "gzip")
}

// isCompressUsingForResponse определяет, можно ли использовать gzip-сжатие для ответа.
// Проверяет:
// - Поддерживает ли клиент gzip (`Accept-Encoding`).
// - Является ли тип содержимого текстовым или JSON.
func isCompressUsingForResponse(w http.ResponseWriter, req *http.Request) bool {
	contentType := req.Header.Get("Content-Type")
	supportsContentType := strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/json")
	acceptEncoding := req.Header.Get("Accept-Encoding")
	supportsGzip := strings.Contains(acceptEncoding, "gzip")
	return supportsGzip && supportsContentType
}
