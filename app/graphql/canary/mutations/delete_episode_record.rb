# frozen_string_literal: true

module Canary
  module Mutations
    class DeleteEpisodeRecord < Canary::Mutations::Base
      argument :episode_record_id, ID, required: true

      field :episode, Canary::Types::Objects::EpisodeType, null: true

      def resolve(episode_record_id:)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        episode_record = context[:viewer].episode_records.published.find_by_graphql_id(episode_record_id)
        episode_record.record.destroy

        {
          episode: episode_record.episode
        }
      end
    end
  end
end
