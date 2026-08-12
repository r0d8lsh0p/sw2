# Build stage — builds from the local source tree (cgo required by lmdb)
FROM golang:1.25-bookworm AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /sw2 .

# Runtime stage
FROM debian:bookworm-slim

WORKDIR /app

# Install iputils and curl
RUN apt-get update && apt-get install -y iputils-ping curl && rm -rf /var/lib/apt/lists/*

COPY --from=builder /sw2 /app/sw2

CMD ["/app/sw2"]
