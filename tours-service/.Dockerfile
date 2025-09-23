# --- FAZA 1: BUILDER ---
FROM golang:1.25.1-alpine AS builder

WORKDIR /app

# 1. KORAK: Kopiramo ceo projekat (build context) u /app
# Ovako preslikavamo celu strukturu projekta u kontejner.
COPY . .

# 2. KORAK: Menjamo radni direktorijum na direktorijum servisa
# OVO JE KLJUČNA ISPRAVKA!
# Sve naredne komande će se izvršavati iz /app/tours-service
WORKDIR /app/tours-service

# 3. KORAK: Skidamo zavisnosti
# Go će sada koristiti `tours-service/go.mod`, a `replace ../common`
# će ispravno pokazivati na `/app/common`.
RUN go mod download
RUN go mod tidy

# 4. KORAK: Kompajliramo aplikaciju
# Putanja `./cmd/server/main.go` je sada ispravna jer smo u `/app/tours-service`.
# Binarni fajl smeštamo u root kontejnera da ga lakše nađemo kasnije.
RUN CGO_ENABLED=0 GOOS=linux go build -o /tours-service-binary ./cmd/server/main.go


# --- FAZA 2: FINAL ---
FROM scratch

# Kopiramo SAMO kompajlirani program iz 'builder' faze.
COPY --from=builder /tours-service-binary .

EXPOSE 8080
EXPOSE 50051

# Pokrećemo binarni fajl
CMD ["./tours-service-binary"]