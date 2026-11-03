defmodule Discordia.EventsContract.AdminTest do
  use ExUnit.Case, async: true

  alias Discordia.EventsContract.Admin

  @secret "test-secret"
  @fraction "2026-10-14T12:30:45.123456Z"
  @whole "2026-10-14T12:30:45Z"

  # {routing_key, event_id, clave del sujeto, id del sujeto, version, occurred_at, firma}.
  # Son los mismos de admin/vectors_test.go; los de servidor se calcularon con un
  # script aparte (HMAC-SHA256 en Python).
  @vectors [
    {"admin.user_suspended", "evt-123", "user_id", "user-42", 3, @fraction,
     "4e4e1885838edbda6e1964eedc9452fd5825874a6c280dba7f73a5f6400ba710"},
    {"admin.user_suspended", "evt-124", "user_id", "user-42", 4, @whole,
     "f96b97d150be7091270411377493ba9ace9b58bfc2f782fd674f02fe639f3f16"},
    {"admin.user_reactivated", "evt-123", "user_id", "user-42", 3, @fraction,
     "79d9f8177d1a8f643952e70bdf2fde24dc2c891783eb36cb216c51a1b3f3649a"},
    {"admin.server_suspended", "evt-123", "server_id", "server-42", 3, @fraction,
     "456c6e2a80ddba4e13664b26baac16f713e2cddea6af70435b522c66a495ca75"},
    {"admin.server_suspended", "evt-124", "server_id", "server-42", 4, @whole,
     "a80f65168feb3ed64c2fa18290c868006fc4bb4a5445572c1e38d0880ba07468"},
    {"admin.server_reactivated", "evt-123", "server_id", "server-42", 3, @fraction,
     "b081fcb4de1d6814961bfc77fa1d575058476adac3e057ca39e15dff0d52e30a"},
    {"admin.server_deletion_requested", "evt-123", "server_id", "server-42", 3, @fraction,
     "d21e2d8f886f4d176aa0a6b212737e305a7b63ee87a602235b131aa540e52bf5"}
  ]

  @routing_keys @vectors |> Enum.map(&elem(&1, 0)) |> Enum.uniq()

  defp event({_routing_key, event_id, key, id, version, occurred_at, signature}) do
    %{
      "event_id" => event_id,
      "occurred_at" => occurred_at,
      key => id,
      "version" => version,
      "signature" => signature
    }
  end

  defp subject_key("admin.user_" <> _), do: "user_id"
  defp subject_key("admin.server_" <> _), do: "server_id"

  describe "vectores fijos" do
    test "sign/6 reproduce cada vector" do
      for {routing_key, event_id, _key, id, version, occurred_at, signature} <- @vectors do
        assert Admin.sign(@secret, routing_key, event_id, id, version, occurred_at) == signature
      end
    end

    test "verify/3 acepta cada vector" do
      for {routing_key, _, _, _, _, _, _} = vector <- @vectors do
        assert Admin.verify(routing_key, event(vector), @secret), routing_key
      end
    end
  end

  describe "verify/3 rechaza" do
    test "cualquier campo firmado alterado" do
      for {routing_key, _, key, _, _, _, _} = vector <- @vectors do
        valid = event(vector)

        for {field, value} <- [
              {"event_id", "evt-999"},
              {key, "otro-id"},
              {"version", valid["version"] + 1},
              {"occurred_at", "2026-10-14T12:30:46Z"}
            ] do
          refute Admin.verify(routing_key, Map.put(valid, field, value), @secret),
                 "#{routing_key} con #{field} alterado"
        end
      end
    end

    test "un secreto distinto, vacío o nil" do
      for {routing_key, _, _, _, _, _, _} = vector <- @vectors do
        refute Admin.verify(routing_key, event(vector), "otro-secreto")
        refute Admin.verify(routing_key, event(vector), "")
        refute Admin.verify(routing_key, event(vector), nil)
      end
    end

    test "una firma vacía, nil, truncada, no hex o en mayúsculas" do
      for {routing_key, _, _, _, _, _, signature} = vector <- @vectors do
        valid = event(vector)

        for bad <- [
              "",
              nil,
              "abcd",
              "no-es-hex",
              binary_part(signature, 0, 63),
              String.upcase(signature)
            ] do
          refute Admin.verify(routing_key, Map.put(valid, "signature", bad), @secret),
                 "#{routing_key} con firma #{inspect(bad)}"
        end
      end
    end

    test "campos faltantes o de tipo equivocado" do
      for {routing_key, _, key, _, _, _, _} = vector <- @vectors do
        valid = event(vector)

        for field <- ["event_id", "occurred_at", key, "version", "signature"] do
          refute Admin.verify(routing_key, Map.delete(valid, field), @secret),
                 "#{routing_key} sin #{field}"
        end

        refute Admin.verify(routing_key, Map.put(valid, "version", "3"), @secret)
        refute Admin.verify(routing_key, Map.put(valid, key, 42), @secret)
        refute Admin.verify(routing_key, Map.put(valid, "occurred_at", 1_700_000_000), @secret)
      end
    end

    test "una routing key desconocida o que no es un binario" do
      valid = event(hd(@vectors))

      refute Admin.verify("admin.otra_cosa", valid, @secret)
      refute Admin.verify("servers.server_deleted", valid, @secret)
      refute Admin.verify(nil, valid, @secret)
      refute Admin.verify(:atom, valid, @secret)
      refute Admin.verify("admin.user_suspended", "no-es-un-mapa", @secret)
    end

    test "un evento con el tipo de sujeto equivocado para su routing key" do
      for {routing_key, _, key, id, _, _, _} = vector <- @vectors do
        other = if key == "user_id", do: "server_id", else: "user_id"
        wrong = vector |> event() |> Map.delete(key) |> Map.put(other, id)

        refute Admin.verify(routing_key, wrong, @secret),
               "#{routing_key} con #{other} en lugar de #{key}"
      end
    end

    test "la firma de un evento reusada bajo cualquier otra routing key" do
      for {from, _, key, id, version, occurred_at, signature} <- @vectors,
          to <- @routing_keys,
          to != from do
        replay = %{
          "event_id" => "evt-123",
          "occurred_at" => occurred_at,
          subject_key(to) => id,
          "version" => version,
          "signature" => signature
        }

        refute Admin.verify(to, replay, @secret),
               "la firma de #{from} (#{key}) no debe valer para #{to}"
      end
    end
  end

  describe "contrato" do
    test "las routing keys son las documentadas" do
      assert Admin.routing_key_user_suspended() == "admin.user_suspended"
      assert Admin.routing_key_user_reactivated() == "admin.user_reactivated"
      assert Admin.routing_key_server_suspended() == "admin.server_suspended"
      assert Admin.routing_key_server_reactivated() == "admin.server_reactivated"
      assert Admin.routing_key_server_deletion_requested() == "admin.server_deletion_requested"
    end

    test "los structs de servidor llevan server_id y no user_id" do
      for module <- [
            Admin.ServerSuspended,
            Admin.ServerReactivated,
            Admin.ServerDeletionRequested
          ] do
        keys = module.__struct__() |> Map.keys()

        assert :server_id in keys
        refute :user_id in keys
        assert [:event_id, :occurred_at, :signature, :version] -- keys == []
      end
    end
  end
end
