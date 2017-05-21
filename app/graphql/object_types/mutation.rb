# frozen_string_literal: true

ObjectTypes::Mutation = GraphQL::ObjectType.define do
  name "Mutation"

  field :updateStatus, field: Mutations::UpdateStatus.field

  field :createRecord, field: Mutations::CreateRecord.field
  field :updateRecord, field: Mutations::UpdateRecord.field
  field :deleteRecord, field: Mutations::DeleteRecord.field
end
