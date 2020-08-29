# frozen_string_literal: true

module Beta
  module Types
    module Objects
      class Mutation < Beta::Types::Objects::Base
        field :updateStatus, mutation: Beta::Mutations::UpdateStatus

        field :createRecord, mutation: Beta::Mutations::CreateRecord
        field :updateRecord, mutation: Beta::Mutations::UpdateRecord
        field :deleteRecord, mutation: Beta::Mutations::DeleteRecord

        field :createReview, mutation: Beta::Mutations::CreateReview
        field :updateReview, mutation: Beta::Mutations::UpdateReview
        field :deleteReview, mutation: Beta::Mutations::DeleteReview
      end
    end
  end
end
