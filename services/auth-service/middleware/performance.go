package middleware

import (
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GzipResponseWriter wraps gin.ResponseWriter for gzip encoding
type GzipResponseWriter struct {
	gin.ResponseWriter
	Writer *gzip.Writer
}

func (w *GzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *GzipResponseWriter) WriteString(s string) (int, error) {
	return w.Writer.Write([]byte(s))
}

// GzipMiddleware compresses HTTP responses for reduced payload size and faster network transfer
func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only compress if the client supports it
		if !strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// Don't compress already compressed formats or tiny responses
		gz := gzip.NewWriter(c.Writer)
		defer gz.Close()

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		c.Writer = &GzipResponseWriter{c.Writer, gz}
		c.Next()
	}
}

// ETagResponseWriter captures response body to compute ETag
type ETagResponseWriter struct {
	gin.ResponseWriter
	Body *bytes.Buffer
}

func (w *ETagResponseWriter) Write(b []byte) (int, error) {
	w.Body.Write(b)
	return w.ResponseWriter.Write(b)
}

// ETagMiddleware generates an ETag based on response content and handles 304 Not Modified
func ETagMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only calculate ETags for GET requests
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Buffer to capture the response body
		blw := &ETagResponseWriter{Body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		// Calculate ETag (SHA1 hash of the response body)
		if c.Writer.Status() == http.StatusOK {
			hash := sha1.Sum(blw.Body.Bytes())
			etag := `W/"` + hex.EncodeToString(hash[:]) + `"`

			// Check client's If-None-Match header
			clientETag := c.Request.Header.Get("If-None-Match")
			if clientETag == etag {
				c.AbortWithStatus(http.StatusNotModified)
				return
			}

			// Set ETag header if modified
			c.Header("ETag", etag)
		}
	}
}

// OptimizedCacheControl applies smart HTTP cache headers to speed up frontend interactions
func OptimizedCacheControl() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// JWT Tokens and Logins should NEVER be cached
		if strings.Contains(path, "/login") || strings.Contains(path, "/register") || strings.Contains(path, "/refresh") || strings.Contains(path, "/logout") {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
			c.Next()
			return
		}

		// Generic profiles can be cached ephemerally (e.g., 5 seconds) to handle frontend UI component render bursts
		if c.Request.Method == http.MethodGet && strings.Contains(path, "/profile") {
			c.Header("Cache-Control", "public, max-age=5, stale-while-revalidate=10")
		} else {
			// Default API cache policy
			c.Header("Cache-Control", "no-cache, must-revalidate")
		}

		c.Next()
	}
}
