# --- FAZA 1: BUILDER ---
FROM golang:1.25.1-alpine AS builder

WORKDIR /app

# 1. KORAK: Kopiramo SVE iz build context-a (ceo projekat)
COPY . .

# 2. KORAK: Menjamo radni direktorijum na direktorijum našeg servisa.
WORKDIR /app/purchase-service

# 3. KORAK: Skidamo zavisnosti
RUN go mod download

# 4. KORAK: Kompajliramo aplikaciju
RUN CGO_ENABLED=0 GOOS=linux go build -o /main ./cmd/main.go


# --- FAZA 2: FINAL ---
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /main .
CMD ["./main"]