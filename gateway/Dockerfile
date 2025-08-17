FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY gateway/go.mod ./

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0 \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0 \
    && apk add --no-cache protobuf

COPY proto-files ./proto-files

# Generate proto files using module-based generation with compatibility options
RUN protoc --proto_path=./proto-files --go_out=. --go_opt=module=gateway \
           --go-grpc_out=. --go-grpc_opt=module=gateway \
           --go-grpc_opt=require_unimplemented_servers=false \
           tours/create-tours.proto

COPY gateway/. .

# Now run go mod tidy after all source code is available
RUN go mod tidy

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


