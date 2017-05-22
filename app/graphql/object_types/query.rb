# frozen_string_literal: true

ObjectTypes::Query = GraphQL::ObjectType.define do
  name "Query"

  field :node, GraphQL::Relay::Node.field
  field :nodes, GraphQL::Relay::Node.plural_field

  connection :searchWorks, ObjectTypes::Work.connection_type do
    argument :annictIds, types[!types.Int]
    argument :seasons, types[!types.String]
    argument :titles, types[!types.String]
    argument :orderBy, InputObjectTypes::WorkOrder

    resolve Resolvers::Works.new
  end
end
