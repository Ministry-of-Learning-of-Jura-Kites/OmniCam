FROM golang:1.24.5 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o omnicam ./backend/main.go

FROM gcr.io/distroless/base-debian11
WORKDIR /app

# Copy binary only
COPY --from=builder /app/omnicam ./

EXPOSE 8080
CMD ["./omnicam"]
