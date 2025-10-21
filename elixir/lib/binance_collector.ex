defmodule BinanceCollector do
  @moduledoc """
  Entry point for the Binance collector application.
  """

  alias BinanceCollector.Collector

  @doc """
  Collects a snapshot for the given `symbol`.
  """
  def collect(symbol, output_dir, opts \\ []) do
    Collector.collect(symbol, output_dir, opts)
  end

  @doc """
  Continuously collect snapshots for `symbol` waiting `interval` milliseconds between runs.

  ## Options

    * `:interval` - interval in milliseconds between fetches (default: 1_000)

  """
  @spec collect_forever(String.t(), Path.t(), keyword()) :: :ok
  def collect_forever(symbol, output_dir, opts \\ []) do
    interval = Keyword.get(opts, :interval, 1_000)

    if interval <= 0 do
      raise ArgumentError, "interval must be greater than zero"
    end

    opts_without_interval = Keyword.delete(opts, :interval)
    do_collect_forever(symbol, output_dir, opts_without_interval, interval)
  end

  defp do_collect_forever(symbol, output_dir, opts, interval) do
    path = collect(symbol, output_dir, opts)
    IO.puts("Snapshot stored in #{path}")

    Process.sleep(interval)
    do_collect_forever(symbol, output_dir, opts, interval)
  end
end
