# Production Dockerfile for Brokle API
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with full optimization
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo -trimpath \
    -ldflags='-w -s -extldflags "-static"' \
    -o bin/brokle \
    ./cmd/server

# Build migration tool
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo -trimpath \
    -ldflags='-w -s -extldflags "-static"' \
    -o bin/migrate \
    ./cmd/migrate

# Final stage - minimal image
FROM scratch

# Copy CA certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binaries
COPY --from=builder /app/bin/brokle /brokle
COPY --from=builder /app/bin/migrate /migrate

# Copy configuration files
COPY --from=builder /app/configs /configs

# Expose port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/brokle"]