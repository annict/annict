# frozen_string_literal: true

module Mutations
  class CreateRecord < Mutations::Base
    argument :episode_id, ID, required: true
    argument :comment, String, required: false
    argument :rating_state, Types::Enums::RatingState, required: false
    argument :share_twitter, Boolean, required: false
    argument :share_facebook, Boolean, required: false

    field :record, Types::Objects::RecordType, null: true

    def resolve(episode_id:, comment: nil, rating_state: nil, share_twitter: nil, share_facebook: nil)
      raise Annict::Errors::InvalidAPITokenScopeError unless context[:doorkeeper_token].writable?

      episode = Episode.only_kept.find_by_graphql_id(episode_id)

      record = episode.episode_records.new do |r|
        r.rating_state = rating_state&.downcase
        r.body = comment
        r.shared_twitter = share_twitter == true
        r.shared_facebook = share_facebook == true
        r.oauth_application = context[:doorkeeper_token].application
      end

      service = NewEpisodeRecordService.new(context[:viewer], record)
      service.ga_client = context[:ga_client]
      service.app = context[:doorkeeper_token].application
      service.via = "graphql_api"

      service.save!

      {
        record: service.episode_record
      }
    end
  end
end
