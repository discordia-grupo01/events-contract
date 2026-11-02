defmodule Discordia.EventsContract.Messages do
  @moduledoc "Reference Elixir binding for message activity events published by messaging."

  @routing_key_message_created "messaging.message_created"
  @routing_key_direct_message_created "messaging.direct_message_created"

  def routing_key_message_created, do: @routing_key_message_created
  def routing_key_direct_message_created, do: @routing_key_direct_message_created

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
      :mentioned_user_ids
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
      :recipient_id
    ]
  end
end
