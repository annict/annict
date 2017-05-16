# frozen_string_literal: true

Types::UserType = GraphQL::ObjectType.define do
  name "User"

  implements GraphQL::Relay::Node.interface

  global_id_field :id
end
