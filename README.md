# Simple With Whitelisting (sw2)

Simple With Whitelisting (sw2) is a nostr relay that displays and accepts notes only from whitelisted pubkeys.

It's built on the [Khatru](https://khatru.nostr.technology) framework.

## Prerequisites

### Docker Setup (Recommended)
- **Docker**: Ensure you have Docker and Docker Compose installed on your system. You can download Docker Desktop from [here](https://www.docker.com/products/docker-desktop/).

### Go Setup (Alternative)
- **Go**: If you prefer to run without Docker, ensure you have Go installed on your system. You can download it from [here](https://golang.org/dl/).
- **Build Essentials**: If you're using Linux, you may need to install build essentials. You can do this by running `sudo apt install build-essential`.

## Use Cases

This relay can suit a variety of uses:

- **Running a small community**: Community members are whitelisted to read and post notes.
- **Running a knowledge base**: Users are whitelisted to read notes, but only administrators can post notes.
- **Running a blind dropbox**: Users are whitelisted to post notes, but only the administrator can read notes.
- **Combinations of the above**: A Community where members can read and post, and guests can read only.

## Setup Instructions

Follow these steps to get the sw2 Relay running on your local machine:

### 1. Clone the repository

```bash
git clone https://github.com/r0d8lsh0p/sw2.git
cd sw2
git checkout docker
```

### 2. Copy `.env.example` to `.env`

You'll need to create an `.env` file based on the example provided in the repository.

```bash
cp .env.example .env
```

### 3. Set your environment variables

Open the `.env` file and set the necessary environment variables. Example variables include:

```bash
RELAY_NAME="utxo's bot relay"
RELAY_PUBKEY="e2ccf7cf20403f3f2a4a55b328f0de3be38558a7d5f33632fdaaefc726c1c8eb"
RELAY_DESCRIPTION="all my bots will use this relay"
RELAY_URL="wss://bots.utxo.one"
RELAY_ICON="https://pfp.nostr.build/d8fb3b6100a0eb9e652bbc34a0c043b7f225dc74e4ed6d733d0e059f9bd444d4.jpg"
RELAY_CONTACT="https://utxo.one"
```

### 4.1 Whitelist Pubkeys for Reading Notes

Open the `read_whitelist.json` file and add pubkeys to the array

```json
{
  "pubkeys": [
    "1c6cb22996baabe921bcd45c8b6213b2dab096f88e4ba5678d43d195a1868551",
    "9c5d0b120f01b75292d2a2bc32972bf918c8dd8927eaa633d3f62e181a292b27",
    "1bda7e1f7396bda2d1ef99033da8fd2dc362810790df9be62f591038bb97c4d9"
  ]
}
```

If the `read_whitelist.json` contains no pubkeys `{"pubkeys": []}`, then all users are authorised to read.

### 4.2 Whitelist Pubkeys for Posting Notes

Open the `write_whitelist.json` file and add pubkeys to the array

```json
{
  "pubkeys": [
    "1c6cb22996baabe921bcd45c8b6213b2dab096f88e4ba5678d43d195a1868551",
    "9c5d0b120f01b75292d2a2bc32972bf918c8dd8927eaa633d3f62e181a292b27",
    "ede41352397758154514148b24112308ced96d121229b0e6a66bc5a2b40c03ec"
  ]
}
```

If the `write_whitelist.json` contains no pubkeys `{"pubkeys": []}`, then all users are authorised to write.

To maintain compatibliity with previous versions of SW2, a file `whitelist.json` can be used instead of `write_whitelist.json` if you prefer.

### 5. Running with Docker (Recommended)

This fork includes Docker support for easy deployment:

```bash
# Start the relay
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the relay
docker-compose down
```

The relay will be accessible at `localhost:3334`.

### 6. Serving with Caddy (Recommended)

This fork is configured to work with Caddy as a reverse proxy. To use Caddy:

1. Uncomment the network configuration in `docker-compose.yml`:
```yaml
networks:
  caddy:
    external: true # Connect to a Caddy network for reverse proxy
```

2. Add a Caddyfile configuration for your domain with WebSocket support:
```
yourdomain.com {
  reverse_proxy sw2-relay:3334 {
    # Enable WebSocket support
    header_up X-Forwarded-For {remote}
    header_up X-Forwarded-Proto {scheme}
    header_up X-Forwarded-Port {server_port}
  }
}
```

3. Ensure your Caddy service is on the correct Docker network.

### 7. Alternative: Build and Run with Go

If you prefer not to use Docker, you can build and run the relay directly:

```bash
go build
./sw2
```

### 8. Alternative: Create a Systemd Service

To have the relay run as a service without Docker, create a systemd unit file:

1. Create the file:
```bash
sudo nano /etc/systemd/system/sw2.service
```

2. Add the following contents:
```ini
[Unit]
Description=sw2 Relay Service
After=network.target

[Service]
ExecStart=/home/ubuntu/sw2/sw2
WorkingDirectory=/home/ubuntu/sw2
Restart=always

[Install]
WantedBy=multi-user.target
```

3. Start and enable the service:
```bash
sudo systemctl daemon-reload
sudo systemctl start sw2
sudo systemctl enable sw2
```

### 9. Access the relay

Once everything is set up, the relay will be running on `localhost:3334` or your domain name if you set up a reverse proxy.

## Upgrading from an older sw2 (database format change)

sw2 now builds on the consolidated [`fiatjaf.com/nostr`](https://pkg.go.dev/fiatjaf.com/nostr) library (the successor to the archived `github.com/fiatjaf/khatru`). **The on-disk database format changed.** A database written by an older sw2 will open without error, but events will not be served (decode errors appear in the logs). Whitelist files, env vars, and read/write behaviour are unchanged, with two exceptions: deletion requests (kind 5) previously bypassed the write whitelist and were not stored — they now pass through the whitelist like any other event and are stored and served (library-driven); and an **empty read whitelist is now publicly readable without authentication**, matching what this README has always said — previously the code demanded NIP-42 auth even with an empty list.

To keep your events, export them with the bundled legacy tool **before** upgrading, using the old database:

```bash
# with the relay stopped
cd tools/export-legacy-db && go build -o export-legacy-db . && cd ../..
./tools/export-legacy-db/export-legacy-db db > events.jsonl

mv db db-old-backup     # start fresh
./sw2 &                  # new binary creates a new-format db/
cat events.jsonl | nak event ws://localhost:3334
```

Replayed events pass the normal write whitelist — so if you have removed
authors from the list since their events were written, those events are
dropped on replay. NIP-70 protected events (`["-"]` tag) cannot be replayed
by the operator at all: the new library only accepts them from their
NIP-42-authenticated author. The export needs RAM proportional to the
database size. If you don't need old events, just move `db/` aside and start.

## Tests

```bash
go test ./...   # unit + in-process integration (the full read/write permission matrix)
```

End-to-end checks spawn the real binary with real whitelist files (the relay listens on the fixed port 3334, so run one mode at a time):

```bash
go build -o sw2 .
go run ./e2e -binary ./sw2 -matrix   # RW / write-only / read-only / neither
go run ./e2e -binary ./sw2 -open     # empty lists: anyone writes, reads are public
go run ./e2e -binary ./sw2 -legacy   # whitelist.json takes primacy over write_whitelist.json
```
