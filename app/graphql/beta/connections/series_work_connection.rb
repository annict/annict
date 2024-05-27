# typed: false
# frozen_string_literal: true

module Beta
  module Connections
    class SeriesWorkConnection < GraphQL::Types::Relay::BaseConnection
      edge_type Beta::Edges::SeriesWorkEdge
    end
  end
end
