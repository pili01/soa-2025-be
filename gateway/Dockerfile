# Start with a builder stage to build the Go application
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 go build -o /app/gateway ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN addgroup -g 1001 -S gateway && \
    adduser -u 1001 -S gateway -G gateway

WORKDIR /app

COPY --from=builder /app/gateway .

RUN chown -R gateway:gateway /app

USER gateway

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./gateway"]