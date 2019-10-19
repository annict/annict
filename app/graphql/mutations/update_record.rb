# frozen_string_literal: true

module Mutations
  class UpdateRecord < Mutations::Base
    argument :record_id, ID, required: true
    argument :comment, String, required: false
    argument :rating_state, Types::Enums::RatingState, required: false
    argument :share_twitter, Boolean, required: false
    argument :share_facebook, Boolean, required: false

    field :record, Types::Objects::RecordType, null: true

    def resolve(record_id:, comment: nil, rating_state: nil, share_twitter: nil, share_facebook: nil)
      raise Annict::Errors::InvalidAPITokenScopeError unless context[:doorkeeper_token].writable?

      record = context[:viewer].episode_records.published.find_by_graphql_id(record_id)

      record.rating_state = rating_state&.downcase
      record.modify_body = record.body != comment
      record.body = comment
      record.shared_twitter = share_twitter == true
      record.shared_facebook = share_facebook == true
      record.oauth_application = context[:doorkeeper_token].application
      record.detect_locale!(:comment)

      record.save!
      record.update_share_record_status
      record.share_to_sns

      {
        record: record
      }
    end
  end
end
