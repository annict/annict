# frozen_string_literal: true

Types::QueryType = GraphQL::ObjectType.define do
  name "Query"

  field :node, GraphQL::Relay::Node.field
  field :nodes, GraphQL::Relay::Node.plural_field

  connection :findWorks, Types::WorkType.connection_type do
    argument :annictIds, types[!types.Int]
    argument :seasons, types[!types.String]
    argument :titles, types[!types.String]

    resolve Resolvers::Works.new
  end
end
