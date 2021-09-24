# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class Mutation < Canary::Types::Objects::Base
        field :updateStatus, mutation: Canary::Mutations::UpdateStatus

        field :skipEpisode, mutation: Canary::Mutations::SkipEpisode

        field :createWorkRecord, mutation: Canary::Mutations::CreateWorkRecord
        field :createEpisodeRecord, mutation: Canary::Mutations::CreateEpisodeRecord
        field :bulkCreateEpisodeRecords, mutation: Canary::Mutations::BulkCreateEpisodeRecords
        field :updateWorkRecord, mutation: Canary::Mutations::UpdateWorkRecord
        field :updateEpisodeRecord, mutation: Canary::Mutations::UpdateEpisodeRecord
        field :deleteRecord, mutation: Canary::Mutations::DeleteRecord

        field :addReaction, mutation: Canary::Mutations::AddReaction
        field :removeReaction, mutation: Canary::Mutations::RemoveReaction

        field :selectProgram, mutation: Canary::Mutations::SelectProgram
        field :unselectProgram, mutation: Canary::Mutations::UnselectProgram
      end
    end
  end
end
