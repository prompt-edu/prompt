package bodyLimit

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testLimit = 1 << 10

type payload struct {
	Value string `json:"value"`
}

func jsonBody(size int) []byte {
	return []byte(`{"value":"` + strings.Repeat("x", size) + `"}`)
}

func newRouter(globalLimit int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Middleware(globalLimit))

	// Mirrors the common handler shape: BindJSON already writes 400 on a failed bind.
	router.POST("/bind", func(c *gin.Context) {
		var p payload
		if err := c.BindJSON(&p); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"length": len(p.Value)})
	})
	// Handlers that answer any bind error with 500 must report 413 as well.
	router.POST("/should-bind", func(c *gin.Context) {
		var p payload
		if err := c.ShouldBindJSON(&p); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"length": len(p.Value)})
	})
	router.POST("/large", WithLimit(4*testLimit), func(c *gin.Context) {
		var p payload
		if err := c.BindJSON(&p); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"length": len(p.Value)})
	})
	router.GET("/get", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	return router
}

func send(router *gin.Engine, method, path string, body []byte) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func TestMiddleware_AllowsBodyUnderLimit(t *testing.T) {
	resp := send(newRouter(testLimit), http.MethodPost, "/bind", jsonBody(100))

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.JSONEq(t, `{"length":100}`, resp.Body.String())
}

func TestMiddleware_RejectsBodyOverLimitWith413(t *testing.T) {
	router := newRouter(testLimit)

	for _, path := range []string{"/bind", "/should-bind"} {
		resp := send(router, http.MethodPost, path, jsonBody(2*testLimit))
		assert.Equal(t, http.StatusRequestEntityTooLarge, resp.Code, path)
	}
}

func TestMiddleware_RejectsChunkedBodyOverLimit(t *testing.T) {
	router := newRouter(testLimit)
	req := httptest.NewRequest(http.MethodPost, "/bind", io.NopCloser(bytes.NewReader(jsonBody(2*testLimit))))
	req.ContentLength = -1 // unknown length, as with chunked transfer encoding
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.Code)
}

func TestWithLimit_RaisesLimitForRoute(t *testing.T) {
	router := newRouter(testLimit)

	resp := send(router, http.MethodPost, "/large", jsonBody(2*testLimit))
	assert.Equal(t, http.StatusOK, resp.Code)

	resp = send(router, http.MethodPost, "/large", jsonBody(5*testLimit))
	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.Code)

	// The raised limit applies to that route only.
	resp = send(router, http.MethodPost, "/bind", jsonBody(2*testLimit))
	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.Code)
}

func TestMiddleware_LeavesRequestsWithoutBodyAlone(t *testing.T) {
	resp := send(newRouter(testLimit), http.MethodGet, "/get", nil)

	assert.Equal(t, http.StatusNoContent, resp.Code)
}

func TestLimitedBody_ReadsExactlyUpToLimit(t *testing.T) {
	body := &limitedBody{source: io.NopCloser(strings.NewReader(strings.Repeat("a", testLimit))), limit: testLimit}

	content, err := io.ReadAll(body)

	require.NoError(t, err)
	assert.Len(t, content, testLimit)
	assert.False(t, body.exceeded)
}

func TestLimitedBody_StopsOneByteOverLimit(t *testing.T) {
	body := &limitedBody{source: io.NopCloser(strings.NewReader(strings.Repeat("a", testLimit+1))), limit: testLimit}

	content, err := io.ReadAll(body)

	var maxBytesErr *http.MaxBytesError
	require.True(t, errors.As(err, &maxBytesErr))
	assert.Equal(t, int64(testLimit), maxBytesErr.Limit)
	assert.Len(t, content, testLimit)
	assert.True(t, body.exceeded)
}
