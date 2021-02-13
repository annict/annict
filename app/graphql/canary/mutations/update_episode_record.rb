# frozen_string_literal: true

module Canary
  module Mutations
    class UpdateEpisodeRecord < Canary::Mutations::Base
      argument :episode_record_id, ID, required: true
      argument :comment, String,
        required: false,
        description: "エピソードへの感想"
      argument :rating_state, Canary::Types::Enums::RatingState,
        required: false,
        description: "エピソードへの評価"
      argument :share_twitter, Boolean,
        required: false,
        description: "エピソードへの記録をTwitterでシェアするかどうか"
      argument :share_facebook, Boolean,
        required: false,
        description: "エピソードへの記録をFacebookでシェアするかどうか"

      field :episode_record, Canary::Types::Objects::EpisodeRecordType, null: true

      def resolve(episode_record_id:, body: nil, rating_state: nil, share_twitter: nil, share_facebook: nil)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        viewer = context[:viewer]
        record = viewer.episode_records.only_kept.find_by_graphql_id(episode_record_id)

        record.rating_state = rating_state&.downcase
        record.modify_body = record.body != body
        record.body = body
        record.oauth_application = context[:application]
        record.detect_locale!(:body)

        record.save!
        viewer.touch(:record_cache_expired_at)

        if share_twitter
          viewer.share_episode_record_to_twitter(record)
        end

        {
          episode_record: record
        }
      end
    end
  end
end
