defmodule Discordia.EventsContract.Servers do
  @moduledoc """
  Elixir binding for the `servers` event domain, kept next to `events.go`
  per this repo's convention. Like `channels/events.ex`, it is the reference
  copy that `messaging` hand-ports into `Messaging.Events.ChannelEvent`;
  the JSON wire shape is the contract.
  """

  @routing_key_server_deleted "servers.server_deleted"

  def routing_key_server_deleted, do: @routing_key_server_deleted

  defmodule ServerDeleted do
    @moduledoc "Published when a server is permanently deleted."

    @type t :: %__MODULE__{
            event_id: String.t(),
            occurred_at: DateTime.t(),
            server_id: String.t(),
            channel_ids: [String.t()]
          }

    defstruct [:event_id, :occurred_at, :server_id, channel_ids: []]
  end
end
