package middleware

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GzipMiddleware handles response compression
func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") ||
			strings.Contains(c.GetHeader("Connection"), "Upgrade") ||
			strings.Contains(c.GetHeader("Content-Type"), "text/event-stream") {
			c.Next()
			return
		}

		gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestSpeed)
		if err != nil {
			c.Next()
			return
		}
		defer gz.Close()

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		c.Writer = &gzipWriter{c.Writer, gz}
		c.Next()
	}
}

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	return g.writer.Write([]byte(s))
}

// ETagMiddleware implements ETag-based caching
func ETagMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// Use a buffer to capture the response body
		bw := &bodyBufferWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = bw

		c.Next()

		// Skip if status is not 200 or body is empty
		if c.Writer.Status() != http.StatusOK || bw.body.Len() == 0 {
			return
		}

		// Calculate ETag
		data := bw.body.Bytes()
		etag := fmt.Sprintf("W/\"%x\"", sha256.Sum256(data))

		c.Header("ETag", etag)
		c.Header("Cache-Control", "no-cache") // Ensure validation with ETag

		if c.GetHeader("If-None-Match") == etag {
			c.Status(http.StatusNotModified)
			c.Writer.Write(nil) // Empty body for 304
			return
		}

		// If not modified, the buffer already contains the data which will be flushed
	}
}

type bodyBufferWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyBufferWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyBufferWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// OptimizedCacheControl adds specific cache headers
func OptimizedCacheControl() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Request.Method == "GET" && c.Writer.Status() == http.StatusOK {
			path := c.Request.URL.Path
			if strings.Contains(path, "/summary") {
				c.Header("Cache-Control", "public, max-age=10") // Short cache for dashboard
			} else if strings.Contains(path, "/slots/available") {
				c.Header("Cache-Control", "public, max-age=5") // Very short cache for slots
			} else if strings.Contains(path, "/reports/") {
				c.Header("Cache-Control", "private, max-age=60") // Report data
			}
		}
	}
}
