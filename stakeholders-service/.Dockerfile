FROM golang:1.22.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 go build -o /app/stakeholders-service ./cmd/server/main.go

FROM scratch

COPY --from=builder /app/stakeholders-service .

EXPOSE 8080

CMD ["./stakeholders-service"]