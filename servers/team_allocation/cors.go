package main

import (
	"strings"

	"github.com/gin-gonic/gin"
	promptSDK "github.com/prompt-edu/prompt-sdk"
)

// multiOriginCORS returns a Gin middleware that supports multiple allowed origins.
//
// The SDK's CORSMiddleware only permits a single origin, which blocks direct
// browser calls from TEASE (Angular dev on :4200, or the dockerised TEASE at
// http://tease.localhost) to this server. This middleware accepts a list of
// origins and echoes back the matching one, which is required because
// Access-Control-Allow-Credentials is true (a wildcard origin is not valid in
// that mode).
//
// Origins are sourced from (in order, deduplicated):
//  1. clientHost (legacy CORE_HOST — kept for back-compat).
//  2. ALLOWED_ORIGINS env var (comma-separated).
//  3. A small dev default covering localhost Angular + docker traefik so local
//     testing works without extra configuration.
func multiOriginCORS(clientHost string) gin.HandlerFunc {
	allowed := buildAllowedOriginSet(clientHost)

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && allowed[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Max-Age", "86400")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "content-Type, content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		}
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	}
}

func buildAllowedOriginSet(clientHost string) map[string]bool {
	allowed := map[string]bool{}

	add := func(raw string) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return
		}
		if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
			raw = "https://" + raw
		}
		allowed[raw] = true
	}

	add(clientHost)

	extra := promptSDK.GetEnv("ALLOWED_ORIGINS", "")
	for _, o := range strings.Split(extra, ",") {
		add(o)
	}

	// Dev defaults for convenient local testing of the TEASE integration.
	// Gated on non-production mode — browsers in prod would never present
	// these origins, but a server-side attacker on the same network could
	// craft a request with Origin: http://localhost and receive a CORS
	// + credentials echo. Gin's release mode is the canonical "prod" signal.
	if !strings.EqualFold(promptSDK.GetEnv("GIN_MODE", ""), "release") {
		add("http://localhost:4200")
		add("http://tease.localhost")
		add("http://localhost")
	}

	return allowed
}
