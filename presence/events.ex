defmodule Discordia.EventsContract.Presence do
  @moduledoc "Reference Elixir binding for presence events published by messaging."

  @routing_key_user_presence_changed "messaging.user_presence_changed"

  def routing_key_user_presence_changed, do: @routing_key_user_presence_changed

  defmodule UserPresenceChanged do
    @enforce_keys [
      :event_id,
      :occurred_at,
      :user_id,
      :connections,
      :do_not_disturb,
      :version
    ]
    defstruct [
      :event_id,
      :occurred_at,
      :user_id,
      :connections,
      :do_not_disturb,
      :version
    ]
  end
end
