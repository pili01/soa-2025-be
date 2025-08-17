FROM golang:1.23.0-alpine AS builder

WORKDIR /app

COPY tours-service/go.mod tours-service/go.sum ./

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest \
    && apk add --no-cache protobuf

RUN go mod tidy

COPY proto-files ./proto-files
RUN ls -la ./proto-files/tours

RUN protoc --proto_path=./proto-files --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           tours/create-tours.proto

COPY tours-service/. .

RUN CGO_ENABLED=0 go build -o /app/tours-service ./cmd/server/main.go

FROM scratch

COPY --from=builder /app/tours-service .

EXPOSE 8081
EXPOSE 50051

CMD ["./tours-service"]