# Stage 1: Build
FROM golang:1.26-alpine AS build

# Install dependencies
RUN apk add --no-cache curl

# Install migrate
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/

# Set the working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 go build -o main .

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
