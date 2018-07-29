# frozen_string_literal: true

module Connections
  class ActivityConnection < GraphQL::Types::Relay::BaseConnection
    edge_type Edges::ActivityEdge
  end
end
