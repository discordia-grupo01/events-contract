defmodule Discordia.EventsContract.Servers do
  @moduledoc """
  Elixir binding for the `servers` event domain, kept next to `events.go`
  per this repo's convention. Like `channels/events.ex`, it is the reference
  copy that `messaging` hand-ports into `Messaging.Events.ChannelEvent`;
  the JSON wire shape is the contract.
  """

  @routing_key_server_deleted "servers.server_deleted"
  @routing_key_emoji_created "servers.emoji_created"
  @routing_key_emoji_deleted "servers.emoji_deleted"

  def routing_key_server_deleted, do: @routing_key_server_deleted
  def routing_key_emoji_created, do: @routing_key_emoji_created
  def routing_key_emoji_deleted, do: @routing_key_emoji_deleted

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

  defmodule EmojiCreated do
    @moduledoc "Published when a custom emoji is added to a server."

    @type t :: %__MODULE__{
            event_id: String.t(),
            occurred_at: DateTime.t(),
            emoji_id: String.t(),
            server_id: String.t(),
            name: String.t(),
            animated: boolean(),
            image_url: String.t()
          }

    defstruct [:event_id, :occurred_at, :emoji_id, :server_id, :name, :image_url, animated: false]
  end

  defmodule EmojiDeleted do
    @moduledoc """
    Published when a custom emoji is deleted. Consumers keep the image so the
    messages and reactions that already use it still show it.
    """

    @type t :: %__MODULE__{
            event_id: String.t(),
            occurred_at: DateTime.t(),
            emoji_id: String.t(),
            server_id: String.t(),
            name: String.t(),
            animated: boolean(),
            image_url: String.t()
          }

    defstruct [:event_id, :occurred_at, :emoji_id, :server_id, :name, :image_url, animated: false]
  end
end
