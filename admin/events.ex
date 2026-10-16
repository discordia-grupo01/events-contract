defmodule Discordia.EventsContract.Admin do
  @moduledoc """
  Elixir binding for the `admin` event domain, kept next to `events.go` per
  this repo's convention. Like the other `events.ex`, it is the reference copy
  that `messaging` and `voice` hand-port; the JSON wire shape and the signature
  are the contract.

  `admin.user_suspended` and `admin.user_reactivated` are signed: admin and
  every consumer share `ADMIN_EVENTS_HMAC_SECRET`, and a consumer drops any
  event whose signature does not verify.

  The signed text is

      HMAC-SHA256(secret, routing_key \\n event_id \\n user_id \\n version \\n occurred_at)

  encoded as lowercase hex. Two details that matter when porting it:

    * `occurred_at` is used **exactly as it arrived in the JSON**. Go formats
      it as RFC 3339 in UTC with the trailing zeros of the fraction trimmed
      (`2026-10-14T12:30:45Z` has no fraction at all), so re-formatting a
      parsed `DateTime` here can change the text and break the signature.
    * The routing key is part of the signed text, so a valid
      `admin.user_suspended` can not be replayed as `admin.user_reactivated`.

  The Go test `admin/vectors_test.go` pins fixed vectors; any port must
  reproduce them.
  """

  @routing_key_user_suspended "admin.user_suspended"
  @routing_key_user_reactivated "admin.user_reactivated"

  def routing_key_user_suspended, do: @routing_key_user_suspended
  def routing_key_user_reactivated, do: @routing_key_user_reactivated

  defmodule UserSuspended do
    @moduledoc "Published when staff suspends an account. Same fields as `UserReactivated`."

    @type t :: %__MODULE__{
            event_id: String.t(),
            occurred_at: String.t(),
            user_id: String.t(),
            version: integer(),
            signature: String.t()
          }

    defstruct [:event_id, :occurred_at, :user_id, :version, :signature]
  end

  defmodule UserReactivated do
    @moduledoc """
    Published when staff lifts a suspension. `version` grows with every
    suspension or reactivation of the same user; a consumer applies an event
    only if its version is higher than the one it already has.
    """

    @type t :: %__MODULE__{
            event_id: String.t(),
            occurred_at: String.t(),
            user_id: String.t(),
            version: integer(),
            signature: String.t()
          }

    defstruct [:event_id, :occurred_at, :user_id, :version, :signature]
  end

  @doc """
  Lowercase hex HMAC for `routing_key` and the JSON fields of an event.
  `occurred_at` must be the string as received.
  """
  @spec sign(String.t(), String.t(), String.t(), String.t(), integer(), String.t()) ::
          String.t()
  def sign(secret, routing_key, event_id, user_id, version, occurred_at)
      when is_binary(secret) and is_binary(routing_key) and is_binary(event_id) and
             is_binary(user_id) and is_integer(version) and is_binary(occurred_at) do
    text =
      Enum.join([routing_key, event_id, user_id, Integer.to_string(version), occurred_at], "\n")

    :crypto.mac(:hmac, :sha256, secret, text) |> Base.encode16(case: :lower)
  end

  @doc """
  Verifies the `signature` of a decoded JSON event (string keys) in constant
  time. An empty secret, a missing, empty, non-hex or upper-case signature, or
  fields of the wrong type never verify.
  """
  @spec verify(String.t(), map(), String.t() | nil) :: boolean()
  def verify(routing_key, %{} = event, secret) when is_binary(routing_key) do
    with true <- is_binary(secret) and secret != "",
         %{
           "event_id" => event_id,
           "occurred_at" => occurred_at,
           "user_id" => user_id,
           "version" => version,
           "signature" => given
         } <- event,
         true <- is_binary(event_id) and is_binary(occurred_at) and is_binary(user_id),
         true <- is_integer(version),
         true <- is_binary(given) and given != "" do
      expected = sign(secret, routing_key, event_id, user_id, version, occurred_at)
      byte_size(given) == byte_size(expected) and :crypto.hash_equals(given, expected)
    else
      _ -> false
    end
  end

  def verify(_routing_key, _event, _secret), do: false
end
