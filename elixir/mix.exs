defmodule BinanceCollector.MixProject do
  use Mix.Project

  def project do
    [
      app: :binance_collector,
      version: "0.1.0",
      elixir: "~> 1.18",
      start_permanent: Mix.env() == :prod,
      deps: []
    ]
  end

  def application do
    [
      extra_applications: [:logger, :inets, :ssl]
    ]
  end
end
