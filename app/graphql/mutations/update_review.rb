# frozen_string_literal: true

module Mutations
  class UpdateReview < Mutations::Base
    argument :review_id, ID, required: true
    argument :title, String, required: false
    argument :body, String, required: true
    AnimeRecord::STATES.each do |state|
      argument state.to_s.camelcase(:lower).to_sym, Types::Enums::RatingState, required: true
    end
    argument :share_twitter, Boolean, required: false
    argument :share_facebook, Boolean, required: false

    field :review, Types::Objects::ReviewType, null: true

    def resolve(
      review_id:,
      title: nil,
      body:,
      rating_overall_state: nil,
      rating_animation_state: nil,
      rating_music_state: nil,
      rating_story_state: nil,
      rating_character_state: nil,
      share_twitter: nil,
      share_facebook: nil
    )
      raise Annict::Errors::InvalidAPITokenScopeError unless context[:doorkeeper_token].writable?

      viewer = context[:viewer]
      work_record = viewer.work_records.only_kept.find_by_graphql_id(review_id)

      work_record.title = title
      work_record.body = body
      AnimeRecord::STATES.each do |state|
        work_record.send("#{state}=".to_sym, send(state.to_s.camelcase(:lower).to_sym)&.downcase)
      end
      work_record.modified_at = Time.now
      work_record.oauth_application = context[:doorkeeper_token].application
      work_record.detect_locale!(:body)

      viewer.setting.attributes = {
        share_review_to_twitter: share_twitter == true,
        share_review_to_facebook: share_facebook == true
      }

      work_record.save!
      viewer.setting.save!

      if share_twitter
        viewer.share_work_record_to_twitter(work_record)
      end

      {
        review: work_record
      }
    end
  end
end
