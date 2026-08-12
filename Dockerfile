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

# Whitelists are read from the working directory at startup; bake the files
# in so image-only deploys work (local dev mounts them as volumes instead,
# and WRITE/READ_WHITELIST_PUBKEYS env vars override them when set).
COPY write_whitelist.json read_whitelist.json /app/

CMD ["/app/sw2"]
