# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class Mutation < Canary::Types::Objects::Base
        field :statusUpdate, mutation: Canary::Mutations::StatusUpdate

        field :episodeRecordCreate, mutation: Canary::Mutations::EpisodeRecordCreate
        field :episodeRecordUpdate, mutation: Canary::Mutations::EpisodeRecordUpdate
        field :episodeRecordDelete, mutation: Canary::Mutations::EpisodeRecordDelete

        field :workRecordCreate, mutation: Canary::Mutations::WorkRecordCreate
        field :workRecordUpdate, mutation: Canary::Mutations::WorkRecordUpdate
        field :workRecordDelete, mutation: Canary::Mutations::WorkRecordDelete

        field :likeWorkRecord, mutation: Canary::Mutations::LikeWorkRecord
        field :unlikeWorkRecord, mutation: Canary::Mutations::UnlikeWorkRecord
      end
    end
  end
end
