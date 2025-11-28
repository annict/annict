# typed: false
# frozen_string_literal: true

module Beta
  module Connections
    class ActivityConnection < GraphQL::Types::Relay::BaseConnection
      edge_type Beta::Edges::ActivityEdge
    end
  end
end
