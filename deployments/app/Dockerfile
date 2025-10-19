# Build stage
FROM golang:1.25-alpine AS base

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

FROM builder AS dev

CMD ["go", "run", "./cmd/server", "-config", "/app/config/config.yaml"]

FROM base as builder

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /binance-live ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /binance-live .

# Copy configuration and migrations
COPY --chown=appuser:appuser config/ ./config/
COPY --chown=appuser:appuser migrations/ ./migrations/

# Switch to non-root user
USER appuser

# Expose port (if needed for future API)
EXPOSE 8080

# Run the application
ENTRYPOINT ["/app/binance-live"]
CMD ["-config", "/app/config/config.yaml"]

