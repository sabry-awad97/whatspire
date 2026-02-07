# syntax=docker/dockerfile:1

# ============================================================================
# Build Stage: Compile the Go application
# ============================================================================
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    gcc \
    musl-dev \
    sqlite-dev

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies (cached if go.mod/go.sum unchanged)
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
# CGO_ENABLED=1 is required for sqlite
# -ldflags flags:
#   -s: strip debug information
#   -w: strip DWARF symbol table
#   -extldflags '-static': create static binary
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.Version=2.0.0" \
    -a \
    -installsuffix cgo \
    -o /build/whatsapp-service \
    ./cmd/whatsapp/main.go

# Verify the binary was created
RUN ls -lh /build/whatsapp-service

# ============================================================================
# Runtime Stage: Create minimal production image
# ============================================================================
FROM alpine:3.23

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    sqlite-libs \
    && update-ca-certificates

# Create non-root user for security
RUN addgroup -g 1000 whatsapp && \
    adduser -D -u 1000 -G whatsapp whatsapp

# Set working directory
WORKDIR /app

# Create necessary directories with proper permissions
RUN mkdir -p /data/media && \
    chown -R whatsapp:whatsapp /app /data

# Copy binary from builder
COPY --from=builder --chown=whatsapp:whatsapp /build/whatsapp-service /app/whatsapp-service

# Copy configuration examples
COPY --chown=whatsapp:whatsapp config.example.yaml /app/config.example.yaml
COPY --chown=whatsapp:whatsapp config.example.json /app/config.example.json

# Switch to non-root user
USER whatsapp

# Expose application port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Set environment variables
ENV GIN_MODE=release \
    TZ=UTC

# Volume for persistent data
VOLUME ["/data"]

# Run the application
ENTRYPOINT ["/app/whatsapp-service"]
