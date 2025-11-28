# typed: false
# frozen_string_literal: true

module Canary
  module Mutations
    class SkipEpisode < Canary::Mutations::Base
      argument :episode_id, ID,
        required: true

      field :episode, Canary::Types::Objects::EpisodeType, null: true
      field :errors, [Canary::Types::Objects::ClientErrorType], null: false

      def resolve(episode_id:)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        viewer = context[:viewer]
        episode = Episode.only_kept.find_by_graphql_id(episode_id)
        library_entry = viewer.library_entries.where(work_id: episode.work_id).first_or_create!

        library_entry.append_episode!(episode)

        {
          episode: episode,
          errors: []
        }
      end
    end
  end
end
