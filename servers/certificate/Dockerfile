# Stage 1: Build Go binary
FROM golang:1.27-alpine AS builder

# Install dependencies
RUN apk add --no-cache curl xz

# Install migrate
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/

# Download pre-built Typst binary (musl-linked, no libc dependency)
ARG TYPST_VERSION=v0.14.2
ARG TARGETARCH=amd64
RUN if [ "$TARGETARCH" = "arm64" ]; then \
      TYPST_ARCH="aarch64-unknown-linux-musl"; \
    else \
      TYPST_ARCH="x86_64-unknown-linux-musl"; \
    fi && \
    curl -L "https://github.com/typst/typst/releases/download/${TYPST_VERSION}/typst-${TYPST_ARCH}.tar.xz" | tar xJ --strip-components=1 -C /usr/local/bin/

# Pin the Go cache locations to the toolchain defaults, so the BuildKit cache
# mounts below provably target the directories `go` actually writes to. The
# mount IDs match servers/Dockerfile, so every Go service shares one cache.
ENV GOMODCACHE=/go/pkg/mod
ENV GOCACHE=/root/.cache/go-build

WORKDIR /app
COPY go.mod go.sum ./
# Requires BuildKit (Docker 23+).
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
# The module cache mount has to be repeated here: cache mounts are not part of
# the image, so without it `go build` would re-download.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o certificate-service

# Stage 2: Final (distroless)
FROM gcr.io/distroless/static-debian12

WORKDIR /app

# Copy the built binary and necessary files
COPY --from=builder /app/certificate-service /app/certificate-service
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate
COPY --from=builder /usr/local/bin/typst /usr/local/bin/typst
COPY --from=builder /app/db/migration /app/db/migration

EXPOSE 8080

CMD ["/app/certificate-service"]
