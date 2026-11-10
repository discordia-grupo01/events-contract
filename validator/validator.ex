defmodule Discordia.EventsContract.Validator do
  @moduledoc """
  Valida el cuerpo de un evento contra el JSON Schema que genera `events-contract`
  a partir de las structs de Go (carpeta `schemas/`). La clave es el routing key
  (`servers.member_left`) o, para el canal interno de `messaging`,
  `broadcast.<evento>`.
  """

  @schemas_dir Path.expand("../schemas", __DIR__)
  @paths Path.wildcard(Path.join(@schemas_dir, "*.json"))
  @uuid ~r/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i

  for path <- @paths, do: @external_resource(path)

  @raw Map.new(@paths, fn path ->
         {Path.basename(path, ".json"), path |> File.read!() |> Jason.decode!()}
       end)

  @spec routing_keys() :: [String.t()]
  def routing_keys, do: @raw |> Map.keys() |> Enum.sort()

  @spec validate(String.t(), map() | binary()) :: :ok | {:error, :unknown_event | [String.t()]}
  def validate(routing_key, payload) when is_binary(payload) do
    case Jason.decode(payload) do
      {:ok, decoded} -> validate(routing_key, decoded)
      {:error, _} -> {:error, ["el cuerpo no es JSON válido"]}
    end
  end

  def validate(routing_key, payload) when is_map(payload) do
    case schema(routing_key) do
      nil ->
        {:error, :unknown_event}

      schema ->
        case ExJsonSchema.Validator.validate(schema, normalize(payload)) do
          :ok -> :ok
          {:error, errors} -> {:error, Enum.map(errors, &format_error/1)}
        end
    end
  end

  @doc false
  def format("uuid", value) when is_binary(value), do: Regex.match?(@uuid, value)
  def format(_format, _value), do: true

  defp normalize(payload), do: payload |> Jason.encode!() |> Jason.decode!()

  defp format_error({message, path}), do: "#{path}: #{message}"

  defp schema(routing_key) do
    case Map.fetch(@raw, routing_key) do
      {:ok, raw} ->
        key = {__MODULE__, routing_key}

        case :persistent_term.get(key, nil) do
          nil ->
            resolved =
              ExJsonSchema.Schema.resolve(raw, custom_format_validator: {__MODULE__, :format})

            :persistent_term.put(key, resolved)
            resolved

          resolved ->
            resolved
        end

      :error ->
        nil
    end
  end
end
