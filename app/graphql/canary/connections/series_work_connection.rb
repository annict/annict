# frozen_string_literal: true

module Connections
  class SeriesWorkConnection < GraphQL::Types::Relay::BaseConnection
    edge_type Edges::SeriesWorkEdge
  end
end
