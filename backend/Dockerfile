FROM golang:1.24.5 AS builder
WORKDIR /app

RUN mkdir -p ./backend/pkg/messages

RUN apt-get update \
 && apt-get install -y --no-install-recommends protobuf-compiler \
 && go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.10 \
 && go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0

COPY go.mod go.sum ./
RUN go mod download

COPY protobufs/ ./protobufs/
RUN protoc --go_out=backend/pkg/messages \
    --go_opt=paths=source_relative protobufs/*.proto

COPY db/ ./db/
COPY sqlc.yaml ./
RUN sqlc generate

COPY ./backend ./backend

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o omnicam ./backend/main.go

FROM gcr.io/distroless/base-debian11
WORKDIR /app

# Copy binary only
COPY --from=builder /app/omnicam ./

EXPOSE 8080
CMD ["./omnicam"]
