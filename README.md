# collector

This repository contains a small Go service that polls Binance for product data every second and stores the raw JSON responses on disk.

## Prerequisites

- [Go 1.22](https://go.dev/dl/)

## Usage

```bash
# build the collector
$ go build ./cmd/binance_collector

# run with default settings (BTCUSDT, 1 second interval)
$ ./binance_collector

# override the symbol or poll interval
$ ./binance_collector -symbol ETHUSDT -interval 2
```

Raw payloads are written to the `data/raw` directory. File names are timestamped in UTC with nanosecond precision so multiple samples per second are retained.

Press `Ctrl+C` to stop the collector gracefully.
