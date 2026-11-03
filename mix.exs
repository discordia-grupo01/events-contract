defmodule EventsContract.MixProject do
  @moduledoc """
  Proyecto mínimo, sin dependencias, para compilar y probar los bindings
  Elixir (`*/events.ex`) que viven al lado de los `events.go`. Go no lo usa.
  """

  use Mix.Project

  def project do
    [
      app: :events_contract,
      version: "0.0.0",
      elixir: "~> 1.16",
      elixirc_paths:
        ~w(admin billing channels membership messages permissions presence servers voice),
      test_paths: ["admin"],
      start_permanent: false,
      deps: []
    ]
  end

  def application, do: [extra_applications: [:crypto]]
end
