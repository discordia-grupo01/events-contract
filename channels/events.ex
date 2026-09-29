defmodule Discordia.EventsContract.Channels do
  @moduledoc """
  Elixir binding for the `channels` event domain, kept next to `events.go`
  per this repo's own convention (one folder per domain, binding lives
  beside the `.go` file it mirrors -- see the README).

  This repo has no Elixir toolchain (no mix.exs, no CI for it), so this
  module is not compiled or tested here -- it is the reference copy that
  `messaging` (the actual Elixir consumer) hand-ports into its own
  `Messaging.Events.ChannelEvent` module. Keep both in sync manually; the
  JSON wire shape is the contract, verified by messaging's own test suite.
  """

  @routing_key_channel_created "servers.channel_created"
  @routing_key_channel_deleted "servers.channel_deleted"

  def routing_key_channel_created, do: @routing_key_channel_created
  def routing_key_channel_deleted, do: @routing_key_channel_deleted

  defmodule ChannelCreated do
    @moduledoc "Published when a text or voice channel is created inside a server."

    @type t :: %__MODULE__{
            event_id: String.t(),
            occurred_at: DateTime.t(),
            server_id: String.t(),
            channel_id: String.t(),
            name: String.t(),
            kind: String.t()
          }

    defstruct [:event_id, :occurred_at, :server_id, :channel_id, :name, :kind]
  end

  defmodule ChannelDeleted do
    @moduledoc "Published when a channel is removed from a server."

    @type t :: %__MODULE__{
            event_id: String.t(),
            occurred_at: DateTime.t(),
            server_id: String.t(),
            channel_id: String.t()
          }

    defstruct [:event_id, :occurred_at, :server_id, :channel_id]
  end
end
