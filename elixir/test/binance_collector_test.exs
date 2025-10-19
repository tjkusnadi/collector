defmodule BinanceCollectorTest do
  use ExUnit.Case, async: true

  alias BinanceCollector.Collector

  test "collect writes ndjson with hourly rotation" do
    {:ok, now} = DateTime.new(~D[2024-03-01], ~T[15:30:00], "Etc/UTC")
    payload = "{\"data\":{\"symbol\":\"BTCUSDT\",\"price\":\"12345.67\"}}"
    http_client = fn _url -> {:ok, payload} end

    tmp_dir = Path.join(System.tmp_dir!(), "binance_collector_test_" <> random_string())
    on_exit(fn -> File.rm_rf!(tmp_dir) end)

    path = Collector.collect("btcusdt", tmp_dir, now: now, http_client: http_client)

    expected = Path.join(tmp_dir, "BTCUSDT-2024030115.ndjson")
    assert path == expected

    [line] =
      path
      |> File.read!()
      |> String.split("\n", trim: true)

    assert line =~ "\"symbol\":\"BTCUSDT\""
    assert line =~ "\"collected_at\":\"#{DateTime.to_iso8601(now)}\""
    assert line =~ "\"payload\":#{payload}"
  end

  defp random_string do
    :crypto.strong_rand_bytes(4)
    |> Base.encode16(case: :lower)
  end
end
