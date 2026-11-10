defmodule Discordia.EventsContract.ValidatorTest do
  use ExUnit.Case, async: true

  alias Discordia.EventsContract.Validator

  @examples Path.expand("../examples", __DIR__)

  test "todos los ejemplos generados desde Go cumplen su schema también en Elixir" do
    paths = Path.wildcard(Path.join(@examples, "*.json"))
    assert length(paths) == length(Validator.routing_keys())

    for path <- paths do
      key = Path.basename(path, ".json")
      body = File.read!(path)

      assert Validator.validate(key, body) == :ok,
             "#{key}: #{inspect(Validator.validate(key, body))}"
    end
  end

  test "falta un campo obligatorio" do
    example = load("servers.member_left") |> Map.delete("user_id")
    assert {:error, [message]} = Validator.validate("servers.member_left", example)
    assert message =~ "user_id"
  end

  test "tipo equivocado" do
    example = load("servers.member_left") |> Map.put("server_id", 5)
    assert {:error, [message]} = Validator.validate("servers.member_left", example)
    assert message =~ "server_id"
  end

  test "formato uuid" do
    example = load("servers.member_left") |> Map.put("event_id", "no-es-uuid")
    assert {:error, [message]} = Validator.validate("servers.member_left", example)
    assert message =~ "event_id"
  end

  test "un campo de más no rompe a nadie" do
    example = load("servers.member_left") |> Map.put("campo_nuevo", 1)
    assert Validator.validate("servers.member_left", example) == :ok
  end

  test "claves con átomos y fechas se normalizan como JSON" do
    event = %{
      event_id: "0b6f1c8e-7a2d-4e3f-9c1b-5d4a3e2f1a0b",
      occurred_at: ~U[2026-11-10 15:00:00Z],
      server_id: "6d5c4b3a-2e1f-4a0b-9c8d-7e6f5a4b3c2d",
      user_id: "u"
    }

    assert Validator.validate("servers.member_left", event) == :ok
  end

  test "JSON roto y evento desconocido" do
    assert {:error, ["el cuerpo no es JSON válido"]} =
             Validator.validate("servers.member_left", "{")

    assert {:error, :unknown_event} = Validator.validate("nadie.nada", %{})
  end

  test "el mensaje de error no incluye el valor recibido" do
    example =
      load("messaging.message_created")
      |> Map.put("content_preview", String.duplicate("secreto", 40))

    assert {:error, [message]} = Validator.validate("messaging.message_created", example)
    refute message =~ "secreto"
  end

  defp load(key), do: @examples |> Path.join(key <> ".json") |> File.read!() |> Jason.decode!()
end
