defmodule Discordia.EventsContract.Messages do
  @moduledoc "Reference Elixir binding for message activity events published by messaging."

  @routing_key_message_created "messaging.message_created"
  @routing_key_direct_message_created "messaging.direct_message_created"

  def routing_key_message_created, do: @routing_key_message_created
  def routing_key_direct_message_created, do: @routing_key_direct_message_created

  @routing_key_message_updated "messaging.message_updated"
  @routing_key_message_deleted "messaging.message_deleted"

  def routing_key_message_updated, do: @routing_key_message_updated
  def routing_key_message_deleted, do: @routing_key_message_deleted

  defmodule MessageCreated do
    @enforce_keys [
      :event_id,
      :occurred_at,
      :server_id,
      :channel_id,
      :message_id,
      :author_id,
      :mentioned_user_ids
    ]
    defstruct [
      :event_id,
      :occurred_at,
      :server_id,
      :channel_id,
      :message_id,
      :author_id,
      :mentioned_user_ids,
      :channel_name,
      :content_preview
    ]
  end

  defmodule DirectMessageCreated do
    @enforce_keys [
      :event_id,
      :occurred_at,
      :conversation_id,
      :message_id,
      :author_id,
      :recipient_id
    ]
    defstruct [
      :event_id,
      :occurred_at,
      :conversation_id,
      :message_id,
      :author_id,
      :recipient_id,
      :content_preview
    ]
  end

  defmodule MessageUpdated do
    @enforce_keys [
      :event_id,
      :occurred_at,
      :server_id,
      :channel_id,
      :message_id,
      :author_id,
      :mentioned_user_ids
    ]
    defstruct [
      :event_id,
      :occurred_at,
      :server_id,
      :channel_id,
      :message_id,
      :author_id,
      :mentioned_user_ids,
      :channel_name,
      :content_preview
    ]
  end

  defmodule MessageDeleted do
    @enforce_keys [:event_id, :occurred_at, :channel_id, :message_id, :author_id]
    defstruct [:event_id, :occurred_at, :server_id, :channel_id, :message_id, :author_id]
  end
end
