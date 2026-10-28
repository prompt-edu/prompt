// Package bodyLimit caps the size of incoming request bodies so a single request cannot make
// the server buffer an arbitrarily large payload (e.g. a huge JSON string on an
// unauthenticated route).
//
// Every service in servers/ carries an identical copy of this package; keep them in sync.
package bodyLimit

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	// DefaultMaxBytes is the body cap applied to every route by Middleware.
	DefaultMaxBytes int64 = 2 << 20 // 2 MiB

	// LargeMaxBytes is meant for the few staff-only routes that carry rich text (the shared
	// editor embeds pasted images as base64) or bulk imports.
	LargeMaxBytes int64 = 25 << 20 // 25 MiB
)

type contextKey struct{}

// Middleware caps every request body at maxBytes. The body is read lazily, so nothing is
// buffered before the route's own middleware (e.g. authentication) has run. Once the cap is
// exceeded, any error status the handler writes is turned into 413, so handlers that answer
// a failed bind with 400 or 500 need no changes. Register it before any route so Gin includes
// it in every route's chain.
func Middleware(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil || c.Request.Body == http.NoBody {
			c.Next()
			return
		}

		body := &limitedBody{source: c.Request.Body, limit: maxBytes}
		c.Request.Body = body
		c.Writer = &tooLargeWriter{ResponseWriter: c.Writer, body: body}
		c.Set(contextKey{}, body)
		c.Next()
	}
}

// WithLimit replaces the cap set by Middleware for a single route. Place it before any
// middleware that reads the body. Without Middleware in the chain it does nothing.
func WithLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if value, ok := c.Get(contextKey{}); ok {
			if body, ok := value.(*limitedBody); ok {
				body.limit = maxBytes
			}
		}
		c.Next()
	}
}

// limitedBody mirrors http.MaxBytesReader, but its limit can still be changed by WithLimit
// after the global middleware has wrapped the body.
type limitedBody struct {
	source   io.ReadCloser
	limit    int64
	read     int64
	exceeded bool
	err      error
}

func (b *limitedBody) Read(p []byte) (int, error) {
	if b.err != nil {
		return 0, b.err
	}
	if len(p) == 0 {
		return 0, nil
	}

	remaining := b.limit - b.read
	if remaining < 0 {
		remaining = 0
	}
	// Read one byte past the limit to tell "exactly at the limit" from "over it".
	if int64(len(p)) > remaining+1 {
		p = p[:remaining+1]
	}
	n, err := b.source.Read(p)
	if int64(n) <= remaining {
		b.read += int64(n)
		b.err = err
		return n, err
	}

	b.read += remaining
	b.exceeded = true
	b.err = &http.MaxBytesError{Limit: b.limit}
	return int(remaining), b.err
}

func (b *limitedBody) Close() error {
	return b.source.Close()
}

// tooLargeWriter reports 413 instead of whatever error status the handler chose once the
// body exceeded its limit.
type tooLargeWriter struct {
	gin.ResponseWriter
	body *limitedBody
}

func (w *tooLargeWriter) WriteHeader(code int) {
	if w.body.exceeded && code >= http.StatusBadRequest {
		code = http.StatusRequestEntityTooLarge
	}
	w.ResponseWriter.WriteHeader(code)
}
