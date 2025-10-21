defmodule BinanceCollector.Collector do
  @moduledoc """
  Collects Binance product data and stores it as NDJSON snapshots.
  """

  @api_url "https://www.binance.com/bapi/asset/v2/public/asset-service/product/get-product-by-symbol"

  @doc """
  Fetches the Binance product for `symbol` and appends it to an NDJSON file.
  """
  @spec collect(String.t(), Path.t(), keyword()) :: Path.t()
  def collect(symbol, output_dir, opts \\ []) when is_binary(symbol) do
    upper_symbol = String.upcase(symbol)
    now = Keyword.get(opts, :now, DateTime.utc_now())
    http_client = Keyword.get(opts, :http_client, &default_http_client/1)

    url = "#{@api_url}?symbol=#{upper_symbol}"

    payload_json =
      case http_client.(url) do
        {:ok, binary} -> String.trim(binary)
        {:error, reason} -> raise "failed to fetch Binance data: #{inspect(reason)}"
      end

    entry = build_entry(upper_symbol, now, payload_json)

    :ok = File.mkdir_p(output_dir)
    file_name = "#{upper_symbol}-#{format_hour(now)}.ndjson"
    path = Path.join(output_dir, file_name)

    :ok = File.write(path, entry, [:append])
    path
  end

  defp build_entry(symbol, %DateTime{} = datetime, payload_json) do
    collected_at = DateTime.to_iso8601(datetime)

    "{\"symbol\":\"#{symbol}\",\"collected_at\":\"#{collected_at}\",\"payload\":#{payload_json}}\n"
  end

  defp format_hour(%DateTime{} = datetime) do
    datetime
    |> DateTime.to_naive()
    |> NaiveDateTime.truncate(:second)
    |> then(&"#{pad(&1.year, 4)}#{pad(&1.month, 2)}#{pad(&1.day, 2)}#{pad(&1.hour, 2)}")
  end

  defp pad(int, size) do
    int
    |> Integer.to_string()
    |> String.pad_leading(size, "0")
  end

  defp default_http_client(url) do
    headers = [{~c"accept", ~c"application/json"}, {~c"user-agent", ~c"collector-elixir/1.0"}]

    case :httpc.request(:get, {String.to_charlist(url), headers}, [], []) do
      {:ok, {{_, 200, _}, _headers, body}} -> {:ok, IO.iodata_to_binary(body)}
      {:ok, {{_, status, _}, _headers, _body}} -> {:error, {:status, status}}
      {:error, reason} -> {:error, reason}
    end
  end
end
