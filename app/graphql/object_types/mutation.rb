# frozen_string_literal: true

ObjectTypes::Mutation = GraphQL::ObjectType.define do
  name "Mutation"

  field :createRecord, field: Mutations::CreateRecord.field
end
