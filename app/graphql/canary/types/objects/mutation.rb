# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class Mutation < Canary::Types::Objects::Base
        field :updateStatus, mutation: Canary::Mutations::UpdateStatus

        field :createEpisodeRecord, mutation: Canary::Mutations::CreateEpisodeRecord
        field :updateEpisodeRecord, mutation: Canary::Mutations::UpdateEpisodeRecord
        field :deleteEpisodeRecord, mutation: Canary::Mutations::DeleteEpisodeRecord

        field :createWorkRecord, mutation: Canary::Mutations::CreateWorkRecord
        field :updateWorkRecord, mutation: Canary::Mutations::UpdateWorkRecord
        field :deleteWorkRecord, mutation: Canary::Mutations::DeleteWorkRecord

        field :likeWorkRecord, mutation: Canary::Mutations::LikeWorkRecord
        field :unlikeWorkRecord, mutation: Canary::Mutations::UnlikeWorkRecord
      end
    end
  end
end
