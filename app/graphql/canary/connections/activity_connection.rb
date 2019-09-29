# frozen_string_literal: true

module Canary
  module Connections
    class ActivityConnection < GraphQL::Types::Relay::BaseConnection
      edge_type Canary::Edges::ActivityEdge
    end
  end
end
