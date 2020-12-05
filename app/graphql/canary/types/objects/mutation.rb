# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class Mutation < Canary::Types::Objects::Base
        field :updateStatus, mutation: Canary::Mutations::UpdateStatus

        field :createEpisodeRecord, mutation: Canary::Mutations::CreateEpisodeRecord
        field :updateEpisodeRecord, mutation: Canary::Mutations::UpdateEpisodeRecord
        field :deleteEpisodeRecord, mutation: Canary::Mutations::DeleteEpisodeRecord

        field :bulkCreateEpisodeRecords, mutation: Canary::Mutations::BulkCreateEpisodeRecords

        field :createAnimeRecord, mutation: Canary::Mutations::CreateAnimeRecord
        field :updateAnimeRecord, mutation: Canary::Mutations::UpdateAnimeRecord
        field :deleteAnimeRecord, mutation: Canary::Mutations::DeleteAnimeRecord

        field :addReaction, mutation: Canary::Mutations::AddReaction
        field :removeReaction, mutation: Canary::Mutations::RemoveReaction

        field :checkProgram, mutation: Canary::Mutations::CheckProgram
        field :uncheckProgram, mutation: Canary::Mutations::UncheckProgram
      end
    end
  end
end
