"""Command line interface for the Binance collector."""

from __future__ import annotations

import argparse
import sys
import time
from pathlib import Path

from .binance import collect


def main() -> None:
    parser = argparse.ArgumentParser(description="Collect Binance product snapshots")
    parser.add_argument(
        "symbol", nargs="?", default="BTCUSDT", help="Trading pair symbol"
    )
    parser.add_argument(
        "--output",
        default=Path("data"),
        type=Path,
        help="Output directory for NDJSON files",
    )
    parser.add_argument(
        "--interval",
        type=float,
        default=1.0,
        help="Interval between fetches in seconds (default: 1.0)",
    )
    args = parser.parse_args()

    if args.interval <= 0:
        print("Interval must be greater than zero", file=sys.stderr)
        raise SystemExit(1)

    try:
        while True:
            file_path = collect(args.symbol, args.output)
            print(f"Snapshot stored in {file_path}")
            time.sleep(args.interval)
    except KeyboardInterrupt:
        print("Stopping collector...")


if __name__ == "__main__":
    main()
