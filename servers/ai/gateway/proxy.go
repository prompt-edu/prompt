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
	cancel       context.CancelFunc
	status       int
	firstByteAt  time.Time
	response     bytes.Buffer
	idle         *time.Timer
	idleExpired  atomic.Bool
	transportErr error
}

func (g *Gateway) forward(c *gin.Context, callID uuid.UUID, logosKey string, request calls.Request) {
	callerCtx := c.Request.Context()
	ctx, cancel := context.WithTimeout(callerCtx, totalTimeout)
	defer cancel()
	exchange := &exchange{cancel: cancel}

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
		// ReverseProxy panics with http.ErrAbortHandler when a started stream breaks.
		aborted := recover()
		g.complete(callerCtx, ctx, callID, request.Streamed, exchange, aborted != nil)
		if aborted != nil {
			// gin's recovery would end a broken response cleanly, so the caller could take it as complete.
			if conn, _, err := http.NewResponseController(c.Writer).Hijack(); err == nil {
				_ = conn.Close()
			}
			panic(aborted)
		}
	}()
	proxy.ServeHTTP(c.Writer, c.Request.WithContext(ctx))
}

func (e *exchange) capture(response *http.Response) error {
	e.status = response.StatusCode
	contentType := response.Header.Get("Content-Type")
	response.Header = http.Header{}
	if contentType != "" {
		response.Header.Set("Content-Type", contentType)
	}
	e.idle = time.AfterFunc(idleTimeout, func() {
		e.idleExpired.Store(true)
		e.cancel()
	})
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
		b.exchange.idle.Reset(idleTimeout)
	}
	return n, err
}

func (g *Gateway) complete(callerCtx, proxyCtx context.Context, callID uuid.UUID, streamed bool, e *exchange, aborted bool) {
	if e.idle != nil {
		e.idle.Stop()
	}
	completion := calls.Completion{
		Outcome:     calls.OutcomeSuccess,
		HTTPStatus:  e.status,
		FirstByteAt: e.firstByteAt,
		Response:    e.response.Bytes(),
		Streamed:    streamed,
	}
	switch delivered := e.transportErr == nil && !aborted; {
	case delivered && e.status >= http.StatusBadRequest:
		completion.Outcome, completion.ErrorCode = calls.OutcomeError, "provider_error"
	case delivered:
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
