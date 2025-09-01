# Forgejo MCP Server Dockerfile
FROM golang:1.23-alpine AS builder

# Install git and ca-certificates (needed for Go module downloads)
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -o forgejo-mcp \
    ./cmd

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/forgejo-mcp .

# Copy config file if it exists
COPY --from=builder /app/config.example.yaml ./config.yaml

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port (if needed for web interface)
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ./forgejo-mcp --help || exit 1

# Set default command
ENTRYPOINT ["./forgejo-mcp"]
CMD ["serve"]