# frozen_string_literal: true

module Canary
  module Edges
    class EpisodeEdge < GraphQL::Types::Relay::BaseEdge
      graphql_name "CustomizedEpisodeEdge"

      node_type Canary::Types::Objects::EpisodeType
    end
  end
end
