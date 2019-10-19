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

        record = context[:viewer].episode_records.published.find_by_graphql_id(episode_record_id)

        record.rating_state = rating_state&.downcase
        record.modify_body = record.body != body
        record.body = body
        record.shared_twitter = share_twitter == true
        record.shared_facebook = share_facebook == true
        record.oauth_application = context[:application]
        record.detect_locale!(:body)

        record.save!
        record.update_share_record_status
        record.share_to_sns

        {
          episode_record: record
        }
      end
    end
  end
end
