package gateway

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httputil"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/ai/calls"
	log "github.com/sirupsen/logrus"
)

type exchange struct {
	status       int
	firstByteAt  time.Time
	response     bytes.Buffer
	idle         *time.Timer
	idleTimeout  time.Duration
	idleExpired  atomic.Bool
	transportErr error
}

func (g *Gateway) forward(c *gin.Context, callID uuid.UUID, logosKey string, request calls.Request) {
	callerCtx := c.Request.Context()
	ctx, cancel := context.WithTimeout(callerCtx, g.totalTimeout)
	defer cancel()
	exchange := &exchange{idleTimeout: g.idleTimeout}
	exchange.idle = time.AfterFunc(g.idleTimeout, func() {
		exchange.idleExpired.Store(true)
		cancel()
	})
	if !request.Streamed {
		// A non-streamed answer arrives only once the whole completion is done, so only the total limit applies until then.
		exchange.idle.Stop()
	}

	proxy := &httputil.ReverseProxy{
		Rewrite: func(proxyRequest *httputil.ProxyRequest) {
			proxyRequest.Out.URL = g.providerURL.JoinPath("chat/completions")
			proxyRequest.Out.Host = ""
			// Fresh headers: the caller's token, cookies and X-Prompt-* never reach the provider.
			proxyRequest.Out.Header = http.Header{}
			proxyRequest.Out.Header.Set("Authorization", "Bearer "+logosKey)
			proxyRequest.Out.Header.Set("Content-Type", "application/json")
			if accept := proxyRequest.In.Header.Get("Accept"); accept != "" {
				proxyRequest.Out.Header.Set("Accept", accept)
			}
			proxyRequest.Out.Body = io.NopCloser(bytes.NewReader(request.Body))
			proxyRequest.Out.ContentLength = int64(len(request.Body))
		},
		FlushInterval:  -1,
		ModifyResponse: exchange.capture,
		ErrorHandler:   exchange.fail,
	}

	defer func() {
		recovered := recover()
		g.complete(callerCtx, ctx, callID, request.Streamed, exchange, recovered)
		if recovered == nil {
			return
		}
		// ReverseProxy panics with http.ErrAbortHandler when a started stream breaks; anything else is a bug.
		if recovered != http.ErrAbortHandler {
			panic(recovered)
		}
		// gin's recovery would end a broken response cleanly, so the caller could take it as complete.
		// gin refuses to hijack once body bytes are written, so the connection is taken from the writer it wraps.
		conn, _, err := http.NewResponseController(unwrap(c.Writer)).Hijack()
		if err != nil {
			panic(recovered)
		}
		_ = conn.Close()
	}()
	proxy.ServeHTTP(c.Writer, c.Request.WithContext(ctx))
}

func unwrap(writer http.ResponseWriter) http.ResponseWriter {
	for {
		wrapper, ok := writer.(interface{ Unwrap() http.ResponseWriter })
		if !ok {
			return writer
		}
		writer = wrapper.Unwrap()
	}
}

func (e *exchange) capture(response *http.Response) error {
	e.status = response.StatusCode
	contentType := response.Header.Get("Content-Type")
	response.Header = http.Header{}
	if contentType != "" {
		response.Header.Set("Content-Type", contentType)
	}
	response.Body = &recordingBody{ReadCloser: response.Body, exchange: e}
	return nil
}

func (e *exchange) fail(writer http.ResponseWriter, _ *http.Request, err error) {
	e.transportErr = err
	status := http.StatusBadGateway
	if errors.Is(err, context.DeadlineExceeded) || e.idleExpired.Load() {
		status = http.StatusGatewayTimeout
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_, _ = writer.Write([]byte(`{"error":"the AI provider could not be reached"}`))
}

type recordingBody struct {
	io.ReadCloser
	exchange *exchange
}

func (b *recordingBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if n > 0 {
		if b.exchange.firstByteAt.IsZero() {
			b.exchange.firstByteAt = time.Now()
		}
		b.exchange.response.Write(p[:n])
		b.exchange.idle.Reset(b.exchange.idleTimeout)
	}
	return n, err
}

func (g *Gateway) complete(callerCtx, proxyCtx context.Context, callID uuid.UUID, streamed bool, e *exchange, recovered any) {
	e.idle.Stop()
	completion := calls.Completion{
		Outcome:     calls.OutcomeSuccess,
		HTTPStatus:  e.status,
		FirstByteAt: e.firstByteAt,
		Response:    e.response.Bytes(),
		Streamed:    streamed,
	}
	switch delivered := e.transportErr == nil && recovered == nil; {
	case delivered && e.status >= http.StatusBadRequest:
		completion.Outcome, completion.ErrorCode = calls.OutcomeError, "provider_error"
	case delivered:
	case recovered != nil && recovered != http.ErrAbortHandler:
		completion.Outcome, completion.ErrorCode = calls.OutcomeError, "internal_error"
	case callerCtx.Err() != nil:
		completion.Outcome, completion.ErrorCode = calls.OutcomeCancelled, "caller_disconnected"
	case e.idleExpired.Load():
		completion.Outcome, completion.ErrorCode = calls.OutcomeTimeout, "idle_timeout"
	case errors.Is(proxyCtx.Err(), context.DeadlineExceeded):
		completion.Outcome, completion.ErrorCode = calls.OutcomeTimeout, "total_timeout"
	case e.transportErr != nil:
		completion.Outcome, completion.ErrorCode = calls.OutcomeError, "provider_unreachable"
	default:
		completion.Outcome, completion.ErrorCode = calls.OutcomeError, "stream_interrupted"
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(callerCtx), completionTimeout)
	defer cancel()
	if err := g.calls.Finish(ctx, callID, completion); err != nil {
		log.WithError(err).WithField("callID", callID).Error("Could not complete the AI call record")
	}
}
