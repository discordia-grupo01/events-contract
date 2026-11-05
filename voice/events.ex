defmodule Discordia.EventsContract.Voice do
  @moduledoc "Reference Elixir binding for voice presence events."

  @routing_key_participant_joined "voice.participant_joined"
  @routing_key_participant_left "voice.participant_left"
  @routing_key_private_call "voice.private_call"

  def routing_key_participant_joined, do: @routing_key_participant_joined
  def routing_key_participant_left, do: @routing_key_participant_left
  def routing_key_private_call, do: @routing_key_private_call

  defmodule PrivateCall do
    defstruct [
      :event_id,
      :event_type,
      :occurred_at,
      :call_id,
      :conversation_id,
      :caller_id,
      :recipient_id,
      :room_name,
      :status,
      :reason,
      :ended_by
    ]
  end

  defmodule ParticipantJoined do
    defstruct [:event_id, :event_type, :occurred_at, :server_id, :room_id, :participant_id, :name]
  end

  defmodule ParticipantLeft do
    defstruct [:event_id, :event_type, :occurred_at, :server_id, :room_id, :participant_id, :name]
  end
end
