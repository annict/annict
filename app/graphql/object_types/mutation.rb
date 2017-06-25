# frozen_string_literal: true

ObjectTypes::Mutation = GraphQL::ObjectType.define do
  name "Mutation"

  field :updateStatus, field: Mutations::UpdateStatus.field

  field :createRecord, field: Mutations::CreateRecord.field
  field :updateRecord, field: Mutations::UpdateRecord.field
  field :deleteRecord, field: Mutations::DeleteRecord.field

  field :createReview, field: Mutations::CreateReview.field
  field :updateReview, field: Mutations::UpdateReview.field
  field :deleteReview, field: Mutations::DeleteReview.field
end
