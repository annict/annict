# frozen_string_literal: true

module Canary
  module Mutations
    class CreateEpisodeRecord < Canary::Mutations::Base
      argument :episode_id, ID, required: true
      argument :comment, String, required: false,
        description: "エピソードへの感想"
      argument :rating_state, Canary::Types::Enums::RatingState, required: false,
        description: "エピソードへの評価"
      argument :share_twitter, Boolean, required: false,
        description: "エピソードへの記録をTwitterでシェアするかどうか"
      argument :share_facebook, Boolean, required: false,
        description: "エピソードへの記録をFacebookでシェアするかどうか"

      field :episode_record, Canary::Types::Objects::EpisodeRecordType, null: true

      def resolve(episode_id:, comment: nil, rating_state: nil, share_twitter: nil, share_facebook: nil)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        episode = Episode.published.find_by_graphql_id(episode_id)

        record = episode.episode_records.new do |r|
          r.rating_state = rating_state&.downcase
          r.comment = comment
          r.shared_twitter = share_twitter == true
          r.shared_facebook = share_facebook == true
          r.oauth_application = context[:application]
        end

        service = NewEpisodeRecordService.new(context[:viewer], record)
        service.app = context[:application]
        service.via = context[:via]

        service.save!

        {
          episode_record: service.episode_record
        }
      end
    end
  end
end
