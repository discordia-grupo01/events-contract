defmodule Discordia.EventsContract.Billing do
  @moduledoc """
  Elixir binding for the `billing` event domain, kept next to `events.go` per
  this repo's convention. Like the other `events.ex`, it is the reference copy
  that `messaging` hand-ports; the JSON wire shape and the signature are the
  contract.

  `billing.premium_updated` is signed: billing and every consumer share
  `BILLING_EVENTS_HMAC_SECRET`, and a consumer drops any event whose
  signature does not verify.

  The signed text is

      HMAC-SHA256(secret, routing_key \\n event_id \\n user_id \\n active \\n
                  premium_until \\n renews \\n version \\n occurred_at)

  encoded as lowercase hex. Details that matter when porting it:

    * `active` and `renews` are the strings `"true"` or `"false"`.
    * `premium_until` is `null` in the JSON for users that never paid; it is
      signed as the empty string.
    * `occurred_at` and `premium_until` are used **exactly as they arrived in
      the JSON**. Go sends them in UTC with the trailing zeros of the fraction
      trimmed (`2026-11-03T12:30:45Z` has no fraction at all), so
      re-formatting a parsed `DateTime` here can change the text and break the
      signature.

  The Go test `billing/vectors_test.go` pins fixed vectors; any port must
  reproduce them.
  """

  @routing_key_premium_updated "billing.premium_updated"

  def routing_key_premium_updated, do: @routing_key_premium_updated

  defmodule PremiumUpdated do
    @moduledoc """
    Full snapshot of a user's Premium. Consumers apply `active`; it turns
    false when Premium ends, so they never need a clock. `version` grows with
    every snapshot of the same user; a consumer applies an event only if its
    version is higher than the one it already has.
    """

    @type t :: %__MODULE__{
            event_id: String.t(),
            occurred_at: String.t(),
            user_id: String.t(),
            active: boolean(),
            premium_until: String.t() | nil,
            renews: boolean(),
            version: integer(),
            signature: String.t()
          }

    defstruct [
      :event_id,
      :occurred_at,
      :user_id,
      :active,
      :premium_until,
      :renews,
      :version,
      :signature
    ]
  end

  @doc """
  Lowercase hex HMAC for the JSON fields of a `billing.premium_updated`.
  `occurred_at` and `premium_until` must be the strings as received
  (`premium_until` may be `nil`).
  """
  @spec sign(
          String.t(),
          String.t(),
          String.t(),
          boolean(),
          String.t() | nil,
          boolean(),
          integer(),
          String.t()
        ) :: String.t()
  def sign(secret, event_id, user_id, active, premium_until, renews, version, occurred_at)
      when is_binary(secret) and is_binary(event_id) and is_binary(user_id) and
             is_boolean(active) and (is_binary(premium_until) or is_nil(premium_until)) and
             is_boolean(renews) and is_integer(version) and is_binary(occurred_at) do
    text =
      Enum.join(
        [
          @routing_key_premium_updated,
          event_id,
          user_id,
          to_string(active),
          premium_until || "",
          to_string(renews),
          Integer.to_string(version),
          occurred_at
        ],
        "\n"
      )

    :crypto.mac(:hmac, :sha256, secret, text) |> Base.encode16(case: :lower)
  end

  @doc """
  Verifies the `signature` of a decoded JSON `billing.premium_updated`
  (string keys) in constant time. An empty secret, a missing, empty, non-hex
  or upper-case signature, or fields of the wrong type never verify.
  """
  @spec verify(map(), String.t() | nil) :: boolean()
  def verify(%{} = event, secret) do
    with true <- is_binary(secret) and secret != "",
         %{
           "event_id" => event_id,
           "occurred_at" => occurred_at,
           "user_id" => user_id,
           "active" => active,
           "premium_until" => premium_until,
           "renews" => renews,
           "version" => version,
           "signature" => given
         } <- event,
         true <- is_binary(event_id) and is_binary(occurred_at) and is_binary(user_id),
         true <- is_boolean(active) and is_boolean(renews) and is_integer(version),
         true <- is_binary(premium_until) or is_nil(premium_until),
         true <- is_binary(given) and given != "" do
      expected =
        sign(secret, event_id, user_id, active, premium_until, renews, version, occurred_at)

      byte_size(given) == byte_size(expected) and :crypto.hash_equals(given, expected)
    else
      _ -> false
    end
  end

  def verify(_event, _secret), do: false
end
