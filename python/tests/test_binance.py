"""Tests for the Binance collector."""

from __future__ import annotations

import datetime as dt
import io
import json
import tempfile
import unittest
from pathlib import Path
from typing import Any

from collector.binance import collect


class StubResponse:
    def __init__(self, payload: dict[str, Any], status: int = 200) -> None:
        self._payload = payload
        self.status = status
        self._buffer: io.StringIO | None = None

    def __enter__(self) -> "StubResponse":
        self._buffer = io.StringIO(json.dumps(self._payload))
        return self

    def __exit__(self, *exc_info: object) -> None:  # pragma: no cover
        if self._buffer is not None:
            self._buffer.close()

    def read(self, size: int | None = -1) -> str:  # pragma: no cover
        assert self._buffer is not None
        return self._buffer.read(size)

    def readline(self, size: int | None = -1) -> str:  # pragma: no cover
        assert self._buffer is not None
        return self._buffer.readline(size)

    def __iter__(self):  # pragma: no cover
        assert self._buffer is not None
        return iter(self._buffer)


class StubOpener:
    def __init__(self, payload: dict[str, Any]) -> None:
        self.payload = payload

    def __call__(self, request: Any) -> StubResponse:
        return StubResponse(self.payload)


class CollectTests(unittest.TestCase):
    def test_collect_writes_ndjson(self) -> None:
        now = dt.datetime(2024, 3, 1, 15, 30, 0, tzinfo=dt.timezone.utc)
        payload = {"data": {"symbol": "BTCUSDT", "price": "12345.67"}}
        opener = StubOpener(payload)

        with tempfile.TemporaryDirectory() as tmpdir:
            tmp_path = Path(tmpdir)
            file_path = collect("btcusdt", tmp_path, now=now, opener=opener)

            self.assertEqual(file_path, tmp_path / "BTCUSDT-2024030115.ndjson")

            contents = file_path.read_text(encoding="utf-8").strip().splitlines()
            self.assertEqual(len(contents), 1)

            entry = json.loads(contents[0])
            self.assertEqual(entry["symbol"], "BTCUSDT")
            self.assertEqual(entry["collected_at"], now.isoformat())
            self.assertEqual(entry["payload"], payload["data"])


if __name__ == "__main__":  # pragma: no cover
    unittest.main()
