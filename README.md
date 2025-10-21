# Binance Multi-language Collector

This repository provides Binance product collectors implemented in Go, Node.js, Python, and Elixir. Each collector fetches product details from the Binance public API and appends the snapshot to an hourly rotated NDJSON file. All CLIs poll the API on an interval (default 1 second).

## Prerequisites

Make sure the following toolchain versions (or newer) are installed:

| Language | Version |
| --- | --- |
| Go | 1.25.3 |
| Node.js | 22 |
| Python | 3.9.22 |
| Elixir | 1.19.0 |

## Project Layout

```
go/       # Go module with CLI in cmd/collector
node/     # Node.js package with CLI in bin/collector.js
python/   # Python package exposing a CLI entry point
elixir/   # Elixir application with Collector module
```

## Installing Dependencies

```bash
# Go
cd go
go mod download

# Node.js (installs dependencies locally)
cd ../node
npm install

# Python
cd ../python
python -m venv .venv
source .venv/bin/activate
# No third-party packages are required, but install any optional dependencies here.

# Elixir
cd ../elixir
mix deps.get
```

> **Note:** The collectors rely on the public Binance API. Keep the polling interval at or above one second to avoid rate limiting.

## Running the Collectors

Each CLI accepts a trading symbol, an output directory, and an optional interval argument.

### Go

```bash
cd go
go run ./cmd/collector --symbol BTCUSDT --output ./data --interval 1s
```

### Node.js

```bash
cd node
node ./bin/collector.js BTCUSDT ./data 1000   # interval in milliseconds
```

### Python

```bash
cd python
python -m collector BTCUSDT --output ./data --interval 1.0
```

### Elixir

```bash
cd elixir
mix run -e "BinanceCollector.collect_forever(\"BTCUSDT\", \"./data\", interval: 1_000)"
```

Stop any CLI with `Ctrl+C`. Snapshots are written to hourly rotated NDJSON files in the specified output directory.

## Running Tests

```bash
cd go && go test ./...
cd ../node && node --test
cd ../python && python -m unittest discover tests
cd ../elixir && mix test
```
