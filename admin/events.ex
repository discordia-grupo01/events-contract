defmodule Discordia.EventsContract.Admin do
  @moduledoc """
  Elixir binding for the `admin` event domain, kept next to `events.go` per
  this repo's convention. Like the other `events.ex`, it is the reference copy
  that `messaging` and `voice` hand-port; the JSON wire shape and the signature
  are the contract.

  `admin.user_suspended`, `admin.user_reactivated`, `admin.server_suspended`,
  `admin.server_reactivated` and `admin.server_deletion_requested` are signed:
  admin and every consumer share `ADMIN_EVENTS_HMAC_SECRET`, and a consumer
  drops any event whose signature does not verify.

  The signed text is

      HMAC-SHA256(secret, routing_key \\n event_id \\n subject_id \\n version \\n occurred_at)

  where the subject is the `user_id` of the account events and the `server_id`
  of the server events. It is encoded as lowercase hex. Two details that matter
  when porting it:

    * `occurred_at` is used **exactly as it arrived in the JSON**. Go formats
      it as RFC 3339 in UTC with the trailing zeros of the fraction trimmed
      (`2026-10-14T12:30:45Z` has no fraction at all), so re-formatting a
      parsed `DateTime` here can change the text and break the signature.
    * The routing key is part of the signed text, so a valid
      `admin.user_suspended` can not be replayed as `admin.user_reactivated`,
      nor an account event as a server event.

  `admin/vectors_test.go` and `admin/events_test.exs` pin the same fixed
  vectors; any port must reproduce them.
  """

  @routing_key_user_suspended "admin.user_suspended"
  @routing_key_user_reactivated "admin.user_reactivated"
  @routing_key_staff_member_added "admin.staff_member_added"
  @routing_key_staff_member_removed "admin.staff_member_removed"
  @routing_key_server_suspended "admin.server_suspended"
  @routing_key_server_reactivated "admin.server_reactivated"
  @routing_key_server_deletion_requested "admin.server_deletion_requested"

  def routing_key_user_suspended, do: @routing_key_user_suspended
  def routing_key_user_reactivated, do: @routing_key_user_reactivated
  def routing_key_server_suspended, do: @routing_key_server_suspended
  def routing_key_server_reactivated, do: @routing_key_server_reactivated
  def routing_key_server_deletion_requested, do: @routing_key_server_deletion_requested
  def routing_key_staff_member_added, do: @routing_key_staff_member_added
  def routing_key_staff_member_removed, do: @routing_key_staff_member_removed

  defmodule StaffMemberAdded do
    @type t :: %__MODULE__{
            event_id: String.t(),
            occurred_at: String.t(),
            user_id: String.t(),
            signature: String.t()
          }

    defstruct [:event_id, :occurred_at, :user_id, :signature]
  end

  defmodule StaffMemberRemoved do
    @type t :: %__MODULE__{
            event_id: String.t(),
            occurred_at: String.t(),
            user_id: String.t(),
            signature: String.t()
          }

    defstruct [:event_id, :occurred_at, :user_id, :signature]
  end

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

  defmodule ServerSuspended do
    @moduledoc """
    Published when staff suspends a server. The server keeps its content but
    nobody can use it until it is reactivated. `version` is admin's
    `moderation_version` for that server: it grows with every suspension,
    reactivation or deletion request, and a consumer applies an event only if
    its version is higher than the one it already has.
    """

    @type t :: %__MODULE__{
            event_id: String.t(),
            occurred_at: String.t(),
            server_id: String.t(),
            version: integer(),
            signature: String.t()
          }

    defstruct [:event_id, :occurred_at, :server_id, :version, :signature]
  end

  defmodule ServerReactivated do
    @moduledoc "Published when staff lifts a server suspension. Same fields as `ServerSuspended`."

    @type t :: %__MODULE__{
            event_id: String.t(),
            occurred_at: String.t(),
            server_id: String.t(),
            version: integer(),
            signature: String.t()
          }

    defstruct [:event_id, :occurred_at, :server_id, :version, :signature]
  end

  defmodule ServerDeletionRequested do
    @moduledoc """
    Published when staff asks to delete a server. It is a request, not the
    fact: `servers` runs the deletion and then publishes `servers.server_deleted`.
    Same fields as `ServerSuspended`.
    """

    @type t :: %__MODULE__{
            event_id: String.t(),
            occurred_at: String.t(),
            server_id: String.t(),
            version: integer(),
            signature: String.t()
          }

    defstruct [:event_id, :occurred_at, :server_id, :version, :signature]
  end

  @doc """
  Lowercase hex HMAC for `routing_key` and the JSON fields of an event.
  `subject_id` is the `user_id` or the `server_id`, depending on the event.
  `occurred_at` must be the string as received.
  """
  @spec sign(String.t(), String.t(), String.t(), String.t(), integer(), String.t()) ::
          String.t()
  def sign(secret, routing_key, event_id, subject_id, version, occurred_at)
      when is_binary(secret) and is_binary(routing_key) and is_binary(event_id) and
             is_binary(subject_id) and is_integer(version) and is_binary(occurred_at) do
    text =
      Enum.join(
        [routing_key, event_id, subject_id, Integer.to_string(version), occurred_at],
        "\n"
      )

    :crypto.mac(:hmac, :sha256, secret, text) |> Base.encode16(case: :lower)
  end

  @doc """
  Verifies the `signature` of a decoded JSON event (string keys) in constant
  time. The subject is read from `"user_id"` for `admin.user_*` and from
  `"server_id"` for `admin.server_*`. An unknown routing key, an event that
  carries the other kind of subject, an empty secret, a missing, empty,
  non-hex or upper-case signature, or fields of the wrong type never verify.
  """
  @spec verify(String.t(), map(), String.t() | nil) :: boolean()
  def verify(routing_key, %{} = event, secret) when is_binary(routing_key) do
    with true <- is_binary(secret) and secret != "",
         subject_key when is_binary(subject_key) <- subject_key(routing_key),
         %{
           "event_id" => event_id,
           "occurred_at" => occurred_at,
           "version" => version,
           "signature" => given
         } <- event,
         subject_id = Map.get(event, subject_key),
         true <- is_binary(event_id) and is_binary(occurred_at) and is_binary(subject_id),
         true <- is_integer(version),
         true <- is_binary(given) and given != "" do
      expected = sign(secret, routing_key, event_id, subject_id, version, occurred_at)
      byte_size(given) == byte_size(expected) and :crypto.hash_equals(given, expected)
    else
      _ -> false
    end
  end

  def verify(_routing_key, _event, _secret), do: false

  defp subject_key(@routing_key_user_suspended), do: "user_id"
  defp subject_key(@routing_key_user_reactivated), do: "user_id"
  defp subject_key(@routing_key_server_suspended), do: "server_id"
  defp subject_key(@routing_key_server_reactivated), do: "server_id"
  defp subject_key(@routing_key_server_deletion_requested), do: "server_id"
  defp subject_key(_routing_key), do: nil
end
