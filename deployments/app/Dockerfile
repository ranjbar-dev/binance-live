# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o binance-live ./cmd/server

# Build CLI tool
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o binance-cli ./cmd/cli

# Production stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create app user
RUN addgroup -g 1001 -S app && \
    adduser -u 1001 -S app -G app

# Set working directory
WORKDIR /app

# Copy built binaries from builder stage
COPY --from=builder /app/binance-live .
COPY --from=builder /app/binance-cli .

# Copy configuration files
COPY --from=builder /app/config ./config
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/sql ./sql

# Change ownership of the app directory
RUN chown -R app:app /app

# Switch to non-root user
USER app

# Expose port (if needed for health checks or metrics)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD echo "Health check placeholder - implement based on your needs"

# Default command
CMD ["./binance-live", "-config", "config/config.yaml"]
