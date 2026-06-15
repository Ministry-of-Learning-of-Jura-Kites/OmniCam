FROM golang:1.25.0 AS builder
WORKDIR /app

RUN mkdir -p ./backend/pkg/messages

RUN apt-get update \
 && apt-get install -y --no-install-recommends protobuf-compiler \
 && && rm -rf /var/lib/apt/lists/* \
 && go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.10 \
 && go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
  go mod download

COPY protobufs/ ./protobufs/
RUN mkdir -p /app/backend/pkg/messages/protobufs \
  && cd protobufs \
  && protoc --go_out=/app/backend/pkg/messages/protobufs \
    --go_opt=paths=source_relative *.proto

COPY db/ ./db/
COPY sqlc.yaml ./
RUN sqlc generate

COPY ./backend ./backend

# Build binary
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /app/omnicam ./backend

FROM debian:11-slim
WORKDIR /app

# Copy binary only
COPY --from=builder /app/omnicam ./

EXPOSE 8080
CMD ["./omnicam"]
