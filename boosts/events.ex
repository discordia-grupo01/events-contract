defmodule Discordia.EventsContract.Boosts do
  @moduledoc """
  Elixir binding for the `boosts` event domain. Only `server_boost_updated` is
  here: the signed saga events are between billing and servers, both in Go, and
  messaging only needs the server's level. The JSON wire shape is the contract.
  """

  @routing_key_updated "servers.server_boost_updated"

  def routing_key_updated, do: @routing_key_updated

  defmodule ServerBoostUpdated do
    @moduledoc "A full snapshot of a server's boosts. Apply it only if `version` is higher."

    @type t :: %__MODULE__{
            event_id: String.t(),
            occurred_at: DateTime.t(),
            server_id: String.t(),
            boost_count: non_neg_integer(),
            level: non_neg_integer(),
            version: integer()
          }

    defstruct [:event_id, :occurred_at, :server_id, :boost_count, :level, :version]
  end
end
