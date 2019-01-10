# frozen_string_literal: true

ObjectTypes::Query = GraphQL::ObjectType.define do
  name "Query"

  field :node, GraphQL::Relay::Node.field
  field :nodes, GraphQL::Relay::Node.plural_field

  field :user, ObjectTypes::User do
    argument :username, !types.String

    resolve ->(_obj, args, _ctx) {
      User.published.find_by(username: args[:username])
    }
  end

  field :viewer, ObjectTypes::User do
    resolve ->(_obj, _args, ctx) {
      ctx[:viewer]
    }
  end

  connection :searchWorks, ObjectTypes::Work.connection_type do
    argument :annictIds, types[!types.Int]
    argument :seasons, types[!types.String]
    argument :titles, types[!types.String]
    argument :orderBy, Types::InputObjects::WorkOrder

    resolve Resolvers::Works.new
  end
end
