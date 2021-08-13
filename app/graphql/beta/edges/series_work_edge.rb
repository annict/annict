# frozen_string_literal: true

module Beta
  module Edges
    class SeriesWorkEdge < GraphQL::Types::Relay::BaseEdge
      node_type Beta::Types::Objects::WorkType
      graphql_name "SeriesWorkEdge"

      field :summary, String, null: true
      field :summary_en, String, null: true
      field :node, Beta::Types::Objects::WorkType, null: false

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
