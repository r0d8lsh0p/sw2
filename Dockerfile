# Build stage
FROM golang:latest as builder

WORKDIR /build

RUN go install github.com/bitvora/sw2@latest

# Runtime stage
FROM debian:bookworm-slim

WORKDIR /app

# Install iputils and curl
RUN apt-get update && apt-get install -y iputils-ping curl && rm -rf /var/lib/apt/lists/*

COPY --from=builder /go/bin/sw2 /app/sw2

RUN chmod +x /app/sw2

CMD ["/app/sw2"]
