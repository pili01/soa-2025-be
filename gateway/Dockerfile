# Build stage
FROM golang:1.23-alpine AS builder

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first (better caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Make sure dependencies are in sync with imports
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Final stage
FROM alpine:latest

# Install certificates
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S gateway && \
    adduser -u 1001 -S gateway -G gateway

WORKDIR /app

# Copy the compiled binary from builder
COPY --from=builder /app/main .

# Set ownership
RUN chown -R gateway:gateway /app

# Switch to non-root user
USER gateway

# Expose application port
EXPOSE 8080

# Healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Start the application
CMD ["./main"]
