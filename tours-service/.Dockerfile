FROM golang:1.22.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 go build -o /app/tours-service ./cmd/server/main.go

FROM scratch

COPY --from=builder /app/tours-service  .

EXPOSE 8081

CMD ["./tours-service"]