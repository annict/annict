# frozen_string_literal: true

module Canary
  module Edges
    class SeriesAnimeEdge < GraphQL::Types::Relay::BaseEdge
      node_type Canary::Types::Objects::AnimeType
      graphql_name "SeriesAnimeEdge"

      field :summary, String,
        null: false

      field :summary_en, String,
        null: false

      field :node, Types::Objects::AnimeType,
        null: false

      def summary
        object.node.summary
      end

      def summary_en
        object.node.summary_en
      end

      def node
        object.node.anime
      end
    end
  end
end
