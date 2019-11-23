# frozen_string_literal: true

module Canary
  module Mutations
    class CreateWorkRecord < Canary::Mutations::Base
      argument :work_id, ID,
        required: true
      argument :body, String,
        required: false,
        description: "作品への感想"
      WorkRecord::STATES.each do |state|
        argument state.to_s.camelcase(:lower).to_sym, Canary::Types::Enums::RatingState,
          required: false,
          description: "作品への評価"
      end
      argument :share_twitter, Boolean,
        required: false,
        description: "作品への記録をTwitterでシェアするかどうか"
      argument :share_facebook, Boolean,
        required: false,
        description: "作品への記録をFacebookでシェアするかどうか"

      field :work_record, Canary::Types::Objects::WorkRecordType, null: true

      def resolve(
        work_id:,
        body: nil,
        rating_overall_state: nil,
        rating_animation_state: nil,
        rating_music_state: nil,
        rating_story_state: nil,
        rating_character_state: nil,
        share_twitter: nil,
        share_facebook: nil
      )
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        work = Work.without_deleted.find_by_graphql_id(work_id)

        review = work.work_records.new do |r|
          r.user = context[:viewer]
          r.work = work
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
          share_review_to_facebook: share_facebook == true
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
        service.via = context[:via]
        service.app = context[:application]

        service
      end
    end
  end
end
