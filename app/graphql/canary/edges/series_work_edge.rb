# typed: false
# frozen_string_literal: true

module Canary
  module Edges
    class SeriesWorkEdge < GraphQL::Types::Relay::BaseEdge
      node_type Canary::Types::Objects::WorkType
      graphql_name "SeriesWorkEdge"

      field :summary, String,
        null: false

      field :summary_en, String,
        null: false

      field :node, Types::Objects::WorkType,
        null: false

      def summary
        object.node.summary
      end

      def summary_en
        object.node.summary_en
      end

      def node
        object.node.work
      end
    end
  end
end
