---
name: golang-patterns
description: Deeper idiomatic Go patterns for PROMPT's services — concurrency (errgroup, context, goroutine leaks), error wrapping, and performance idioms. The go/ rules point here for these; they own the Gin/pgx/sqlc conventions.
metadata:
  origin: ECC
---

<!--
Vendored from Everything Claude Code (ECC) — https://github.com/affaan-m/everything-claude-code
License: MIT (Copyright 2026 Affaan Mustafa). Adapted for PROMPT 2.0 — trimmed to the idioms
the `.claude/rules/go/` rules defer to (concurrency, context, errors, performance,
anti-patterns); generic 101 material and a non-PROMPT project layout were removed.
-->

# Go Development Patterns

Deeper Go idioms for PROMPT's backend services. This skill is the reference the go rules point to
for **concurrency, context, error handling, and anti-patterns**. It does **not** restate the
PROMPT-specific conventions those rules own:

- `go/coding-style.md` — stack (Go 1.26, Gin, pgx, sqlc, golang-migrate), formatting, naming,
  module layout (`moduleDTO/`, `router.go`, `service.go`, `validation.go`), logging, env vars.
- `go/patterns.md` — thin Gin handlers, `prompt-sdk`, transactions (`DeferDBRollback`).
- `go/sqlc.md` — DB access through generated sqlc methods (never hand-written SQL in Go).

## When to Activate

- Writing or reviewing concurrent Go (goroutines, channels, cancellation)
- Deciding how to wrap and inspect errors
- Tightening a hot path for allocations

## Error Handling

### Wrap with context, inspect with `errors.Is` / `errors.As`

```go
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("load config %s: %w", path, err)  // %w preserves the chain
    }
    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parse config %s: %w", path, err)
    }
    return &cfg, nil
}
```

```go
// Sentinel errors and typed errors are both inspectable through a wrapped chain.
var ErrNotFound = errors.New("resource not found")

type ValidationError struct{ Field, Message string }
func (e *ValidationError) Error() string { return e.Field + ": " + e.Message }

func handle(err error) {
    if errors.Is(err, ErrNotFound) { /* 404 */ return }
    var ve *ValidationError
    if errors.As(err, &ve) { /* 400 using ve.Field */ return }
}
```

### Don't silently drop errors

```go
result, _ := doSomething()          // Bad: swallows the failure

_ = writer.Close()                  // Acceptable: best-effort cleanup, documented as such
```

## Concurrency

### `errgroup` for coordinated fan-out (cancel-on-first-error)

```go
import "golang.org/x/sync/errgroup"

func FetchAll(ctx context.Context, ids []string) ([]Result, error) {
    g, ctx := errgroup.WithContext(ctx)
    results := make([]Result, len(ids))
    for i, id := range ids {
        i, id := i, id                 // capture loop vars (pre-1.22 habit; harmless after)
        g.Go(func() error {
            r, err := fetch(ctx, id)
            if err != nil {
                return err              // first error cancels ctx for the rest
            }
            results[i] = r
            return nil
        })
    }
    if err := g.Wait(); err != nil {
        return nil, err
    }
    return results, nil
}
```

### Context for timeouts and cancellation

Take `ctx context.Context` as the **first parameter** and thread it through I/O. In handlers use
`c.Request.Context()` so a client disconnect cancels the work.

```go
func FetchWithTimeout(ctx context.Context, url string) ([]byte, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("fetch %s: %w", url, err)
    }
    defer resp.Body.Close()
    return io.ReadAll(resp.Body)
}
```

### Avoid goroutine leaks

A goroutine that sends on an unbuffered channel blocks forever if the receiver is gone. Buffer the
channel or select on `ctx.Done()`:

```go
func safeFetch(ctx context.Context, url string) <-chan []byte {
    ch := make(chan []byte, 1)          // buffered: send never blocks the goroutine
    go func() {
        data, err := fetch(url)
        if err != nil {
            return
        }
        select {
        case ch <- data:
        case <-ctx.Done():
        }
    }()
    return ch
}
```

### Graceful shutdown of the HTTP server

```go
func GracefulShutdown(server *http.Server) {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("forced shutdown: %v", err)
    }
}
```

## Design Idioms

### Make the zero value useful

```go
type Counter struct {
    mu    sync.Mutex   // usable unlocked
    count int          // starts at 0
}
// vs. a struct holding a nil map/slice that panics until initialized.
```

### Functional options for optional config

```go
type Option func(*Server)

func WithTimeout(d time.Duration) Option { return func(s *Server) { s.timeout = d } }

func NewServer(addr string, opts ...Option) *Server {
    s := &Server{addr: addr, timeout: 30 * time.Second}  // defaults
    for _, opt := range opts {
        opt(s)
    }
    return s
}
```

### Inject dependencies; avoid package-level mutable state

```go
// Bad: global set up in init() — hard to test, shared across everything.
var db *sql.DB

// Good: hold dependencies on a struct, pass them in.
type Server struct{ db *pgxpool.Pool }
func NewServer(db *pgxpool.Pool) *Server { return &Server{db: db} }
```

## Performance (measure first)

```go
// Preallocate when the size is known — avoids repeated slice growth.
results := make([]Result, 0, len(items))

// Build strings with strings.Builder (or strings.Join), not += in a loop.
var sb strings.Builder
for i, p := range parts {
    if i > 0 { sb.WriteString(",") }
    sb.WriteString(p)
}

// sync.Pool for hot, short-lived allocations (e.g. per-request buffers).
var bufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}
```

## Anti-Patterns to Avoid

```go
// Naked returns in a long function — unclear what's returned.
func process() (result int, err error) { /* 50 lines */ return }

// panic for ordinary control flow — return an error instead.
func GetUser(id string) *User { u, err := db.Find(id); if err != nil { panic(err) }; return u }

// Storing context.Context in a struct — pass it as the first argument.
type Request struct { ctx context.Context; ID string }

// Mixing value and pointer receivers on one type — pick one and be consistent.
func (c Counter) Value() int  { return c.n }
func (c *Counter) Increment() { c.n++ }
```

**Remember:** clear beats clever; return early and keep the happy path unindented; treat errors as
values. `gofmt`/`goimports` are mandatory (`go/coding-style.md`).

---

*Adapted from ECC's Go patterns (MIT License, copyright Affaan Mustafa).*
