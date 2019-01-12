# frozen_string_literal: true

module Types
  module Objects
    class Mutation < Types::Objects::Base
      field :updateStatus, mutation: Mutations::UpdateStatus

      field :createRecord, mutation: Mutations::CreateRecord
      field :updateRecord, mutation: Mutations::UpdateRecord
      field :deleteRecord, mutation: Mutations::DeleteRecord

      field :createReview, mutation: Mutations::CreateReview
      field :updateReview, mutation: Mutations::UpdateReview
      field :deleteReview, mutation: Mutations::DeleteReview
    end
  end
end
