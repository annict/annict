# frozen_string_literal: true

module Canary
  module Connections
    class SeriesAnimeConnection < GraphQL::Types::Relay::BaseConnection
      edge_type Canary::Edges::SeriesAnimeEdge
    end
  end
end
