# frozen_string_literal: true

module Canary
  module Connections
    class EpisodeConnection < GraphQL::Types::Relay::BaseConnection
      graphql_name "CustomizedEpisodeConnection"

      edge_type Canary::Edges::EpisodeEdge

      field :total_count, Integer, null: false

      def total_count
        object.nodes.size
      end
    end
  end
end
