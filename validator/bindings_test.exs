defmodule Discordia.EventsContract.BindingsTest do
  use ExUnit.Case, async: true

  @schemas Path.expand("../schemas", __DIR__)

  defp schema_keys do
    for path <- Path.wildcard(Path.join(@schemas, "*.json")), into: %{} do
      key = Path.basename(path, ".json")

      props =
        path
        |> File.read!()
        |> Jason.decode!()
        |> Map.fetch!("properties")
        |> Map.keys()
        |> Enum.sort()

      {key, props}
    end
  end

  defp struct_modules do
    {:ok, modules} = :application.get_key(:events_contract, :modules)

    for module <- modules,
        String.starts_with?(Atom.to_string(module), "Elixir.Discordia.EventsContract."),
        function_exported?(module, :__struct__, 0) do
      module
    end
  end

  defp snake(module) do
    module
    |> Module.split()
    |> List.last()
    |> Macro.underscore()
  end

  test "cada struct de los bindings coincide con el schema de su evento" do
    schemas = schema_keys()
    assert struct_modules() != []

    for module <- struct_modules() do
      matches =
        for {key, props} <- schemas,
            String.ends_with?(key, "." <> snake(module)),
            do: {key, props}

      assert matches != [], "#{inspect(module)} no tiene schema"

      fields =
        module
        |> struct()
        |> Map.keys()
        |> List.delete(:__struct__)
        |> Enum.map(&Atom.to_string/1)
        |> Enum.sort()

      assert Enum.any?(matches, fn {_key, props} -> props == fields end),
             "#{inspect(module)} tiene #{inspect(fields)} y el schema #{inspect(Enum.map(matches, &elem(&1, 1)))}"
    end
  end

  test "cada routing key de los bindings existe en el contrato" do
    schemas = schema_keys()
    {:ok, modules} = :application.get_key(:events_contract, :modules)

    keys =
      for module <- modules,
          String.starts_with?(Atom.to_string(module), "Elixir.Discordia.EventsContract."),
          {fun, 0} <- module.__info__(:functions),
          String.starts_with?(Atom.to_string(fun), "routing_key_"),
          do: apply(module, fun, [])

    assert keys != []

    for key <- keys do
      assert Map.has_key?(schemas, key),
             "#{key} está en un binding de Elixir pero no en el contrato"
    end
  end
end
