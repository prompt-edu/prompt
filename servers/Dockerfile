# Stage 1: Build
FROM golang:1.26-alpine AS build

# Install dependencies
RUN apk add --no-cache curl

# Install migrate
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/

# Pin the Go cache locations to the toolchain defaults, so the BuildKit cache
# mounts below provably target the directories `go` actually writes to.
ENV GOMODCACHE=/go/pkg/mod
ENV GOCACHE=/root/.cache/go-build

# Set the working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies. The mount IDs default to the target path, so all
# services built from this file share one module cache instead of each
# downloading its own copy. Requires BuildKit (Docker 23+).
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the source code
COPY . .

# Build the Go app. The module cache mount has to be repeated here: cache mounts
# are not part of the image, so without it `go build` would re-download. The
# build cache is what makes a source-only change recompile just the changed
# package instead of the whole dependency graph.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -o main .

# Stage 2: Final (distroless)
FROM gcr.io/distroless/static-debian12

# Set the working directory
WORKDIR /app

# Copy the built binary and necessary files from the build stage
COPY --from=build /app/main /app/main
COPY --from=build /usr/local/bin/migrate /usr/local/bin/migrate
COPY --from=build /app/db/migration /app/db/migration

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["/app/main"]
