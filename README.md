# collector

This repository contains small Go and Node.js collectors that poll Binance for product data every second and append the raw JSON responses to hourly [NDJSON](https://github.com/ndjson/ndjson-spec) files on disk.

## Prerequisites

- [Go 1.22](https://go.dev/dl/)
- [Node.js 18+](https://nodejs.org/en/download)

## Usage

### Go collector

```bash
# build the collector
$ go build ./cmd/binance_collector

# run with default settings (BTCUSDT, 1 second interval)
$ ./binance_collector

# override the symbol or poll interval
$ ./binance_collector -symbol ETHUSDT -interval 2
```

### Node.js collector

```bash
# run with default settings (BTCUSDT, 1 second interval)
$ node js/bin/binance_collector.js

# override the symbol, interval, or output directory
$ node js/bin/binance_collector.js --symbol ETHUSDT --interval 2 --output ./custom/raw

# collect a finite number of samples (useful for testing)
$ node js/bin/binance_collector.js --samples 5
```

Raw payloads are written to the `data/raw` directory by default. Each trading pair appends to an hourly file named `${SYMBOL}-YYYY-MM-DD-HH.ndjson` (timestamps are in UTC), ensuring every sample is preserved in newline-delimited JSON format.

Press `Ctrl+C` to stop either collector gracefully.
