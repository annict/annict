# frozen_string_literal: true

module Canary
  module Mutations
    class CreateEpisodeRecord < Canary::Mutations::Base
      argument :episode_id, ID,
        required: true
      argument :comment, String,
        required: false,
        description: "エピソードへの感想"
      argument :rating, Canary::Types::Enums::Rating,
        required: false,
        description: "エピソードへの評価"
      argument :share_to_twitter, Boolean,
        required: false,
        description: "エピソードへの記録をTwitterでシェアするかどうか"

      field :record, Canary::Types::Objects::RecordType, null: true
      field :errors, [Canary::Types::Objects::ClientErrorType], null: false

      def resolve(episode_id:, comment: nil, rating: nil, share_to_twitter: nil)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        viewer = context[:viewer]
        episode = Episode.only_kept.find_by_graphql_id(episode_id)

        creator = EpisodeRecordCreator.new(
          user: viewer,
          episode: episode,
          rating: rating,
          comment: comment,
          share_to_twitter: share_to_twitter
        ).call

        {
          record: creator.record,
          errors: creator.errors
        }
      end
    end
  end
end
