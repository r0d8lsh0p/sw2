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

# Whitelist files are read from the working directory at startup and the
# process exits when they are missing, so image-only deploys need them baked
# in (docker-compose mounts the real files over these; WRITE/READ_WHITELIST_
# PUBKEYS env vars override them when set). The baked write whitelist is a
# fail-closed placeholder — a non-empty list whose entry can never match a
# real pubkey — so a deploy that forgets its env var rejects every write
# instead of silently accepting whatever keys the repo copy happens to hold.
# The baked read whitelist is empty = publicly readable, sw2's documented
# default.
RUN echo '{"pubkeys":["set-WRITE_WHITELIST_PUBKEYS-or-mount-write_whitelist.json"]}' > /app/write_whitelist.json
COPY read_whitelist.json /app/

CMD ["/app/sw2"]
