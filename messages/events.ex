defmodule Discordia.EventsContract.Messages do
  @moduledoc "Reference Elixir binding for message activity events published by messaging."

  @routing_key_message_created "messaging.message_created"

  def routing_key_message_created, do: @routing_key_message_created

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
end
