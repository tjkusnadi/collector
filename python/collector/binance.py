"""Utilities for collecting Binance product data."""

from __future__ import annotations

import datetime as dt
import json
import os
import pathlib
import urllib.parse
import urllib.request
from typing import Any, Optional

API_URL = "https://www.binance.com/bapi/asset/v2/public/asset-service/product/get-product-by-symbol"


class BinanceError(RuntimeError):
    """Raised when the Binance API returns an error response."""


def _ensure_symbol(symbol: str) -> str:
    if not symbol:
        raise ValueError("symbol is required")
    return symbol.upper()


def collect(
    symbol: str,
    output_dir: os.PathLike[str] | str,
    *,
    now: Optional[dt.datetime] = None,
    opener: Optional[Any] = None,
) -> pathlib.Path:
    """Collect Binance product information and append it to an NDJSON file."""

    upper_symbol = _ensure_symbol(symbol)
    base_dir = pathlib.Path(output_dir)
    if not base_dir:
        raise ValueError("output_dir is required")

    timestamp = (now or dt.datetime.utcnow()).astimezone(dt.timezone.utc)

    params = urllib.parse.urlencode({"symbol": upper_symbol})
    url = f"{API_URL}?{params}"

    request = urllib.request.Request(
        url,
        headers={
            "Accept": "application/json",
            "User-Agent": "collector-python/1.0",
        },
    )

    opener = opener or urllib.request.urlopen
    with opener(request) as response:
        if response.status != 200:
            raise BinanceError(f"unexpected status: {response.status}")
        payload = json.load(response)

    entry = {
        "symbol": upper_symbol,
        "collected_at": timestamp.isoformat(),
        "payload": payload.get("data", {}),
    }

    base_dir.mkdir(parents=True, exist_ok=True)
    file_name = f"{upper_symbol}-{timestamp.strftime('%Y%m%d%H')}.ndjson"
    file_path = base_dir / file_name

    with file_path.open("a", encoding="utf-8") as f:
        json.dump(entry, f)
        f.write("\n")

    return file_path
