# typed: false
# frozen_string_literal: true

module Canary
  module Connections
    class SeriesWorkConnection < GraphQL::Types::Relay::BaseConnection
      edge_type Canary::Edges::SeriesWorkEdge
    end
  end
end
