defmodule EventsContract.MixProject do
  @moduledoc """
  Bindings Elixir (`*/events.ex`) y validador de payloads contra los JSON Schema
  generados en `schemas/`. Los servicios Elixir lo usan como dependencia de git.
  Go no lo usa.
  """

  use Mix.Project

  def project do
    [
      app: :events_contract,
      version: "0.0.0",
      elixir: "~> 1.16",
      elixirc_paths:
        ~w(admin billing channels membership messages permissions presence servers voice validator),
      test_paths: ["admin", "validator"],
      start_permanent: false,
      deps: deps()
    ]
  end

  def application, do: [extra_applications: [:crypto]]

  defp deps do
    [
      {:ex_json_schema, "~> 0.10"},
      {:jason, "~> 1.4"}
    ]
  end
end
