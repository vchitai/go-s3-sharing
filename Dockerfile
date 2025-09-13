# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/cli ./cmd/cli

# Final stage
FROM alpine:3.18

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy binaries from builder stage
COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/bin/cli /app/cli

# Make binaries executable
RUN chmod +x /app/server /app/cli

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Default command
CMD ["./server"]
