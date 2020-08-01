# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class Mutation < Canary::Types::Objects::Base
        field :updateStatus, mutation: Canary::Mutations::UpdateStatus

        field :createEpisodeRecord, mutation: Canary::Mutations::CreateEpisodeRecord
        field :updateEpisodeRecord, mutation: Canary::Mutations::UpdateEpisodeRecord
        field :deleteEpisodeRecord, mutation: Canary::Mutations::DeleteEpisodeRecord

        field :createAnimeRecord, mutation: Canary::Mutations::CreateAnimeRecord
        field :updateAnimeRecord, mutation: Canary::Mutations::UpdateAnimeRecord
        field :deleteAnimeRecord, mutation: Canary::Mutations::DeleteAnimeRecord

        field :likeAnimeRecord, mutation: Canary::Mutations::LikeAnimeRecord
        field :unlikeAnimeRecord, mutation: Canary::Mutations::UnlikeAnimeRecord
      end
    end
  end
end
