FROM golang:1.23.0-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/tours-service ./cmd/server/main.go

FROM scratch

COPY --from=builder /app/tours-service .

EXPOSE 8080 
EXPOSE 50051 

CMD ["./tours-service"]