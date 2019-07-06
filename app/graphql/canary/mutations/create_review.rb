# frozen_string_literal: true

module Mutations
  class CreateReview < Mutations::Base
    argument :work_id, ID, required: true
    argument :title, String, required: false
    argument :body, String, required: true
    WorkRecord::STATES.each do |state|
      argument state.to_s.camelcase(:lower).to_sym, Types::Enums::RatingState, required: false
    end
    argument :share_twitter, Boolean, required: false
    argument :share_facebook, Boolean, required: false

    field :review, Types::Objects::ReviewType, null: true

    def resolve(
      work_id:,
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
      raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

      work = Work.published.find_by_graphql_id(work_id)

      review = work.work_records.new do |r|
        r.user = context[:viewer]
        r.work = work
        r.title = title
        r.body = body
        r.rating_overall_state = rating_overall_state
        r.rating_animation_state = rating_animation_state
        r.rating_music_state = rating_music_state
        r.rating_story_state = rating_story_state
        r.rating_character_state = rating_character_state
        r.oauth_application = context[:application]
      end
      context[:viewer].setting.attributes = {
        share_review_to_twitter: share_twitter == true,
        share_review_to_facebook: :share_facebook == true
      }

      service = work_record_service(review)
      service.save!

      {
        review: service.work_record
      }
    end

    private

    def work_record_service(review)
      service = NewWorkRecordService.new(context[:viewer], review, context[:viewer].setting)
      service.via = "graphql_api"
      service.app = context[:application]
      service.ga_client = context[:ga_client]

      service
    end
  end
end
