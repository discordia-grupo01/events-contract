defmodule Discordia.EventsContract.Membership do
  @routing_key_member_left "servers.member_left"

  def routing_key_member_left, do: @routing_key_member_left

  defmodule MemberLeft do
    @moduledoc """
    Published when a user leaves a server, is kicked or banned from it, or the
    server is deleted (one per member).
    """

    @type t :: %__MODULE__{
            event_id: String.t(),
            occurred_at: DateTime.t(),
            server_id: String.t(),
            user_id: String.t()
          }

    defstruct [:event_id, :occurred_at, :server_id, :user_id]
  end
end
