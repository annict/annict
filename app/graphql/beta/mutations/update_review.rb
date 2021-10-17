# frozen_string_literal: true

module Beta
  module Mutations
    class UpdateReview < Beta::Mutations::Base
      argument :review_id, ID, required: true
      argument :title, String, required: false
      argument :body, String, required: true
      WorkRecord::STATES.each do |state|
        argument state, Beta::Types::Enums::RatingState, required: true
      end
      argument :share_twitter, Boolean, required: false
      argument :share_facebook, Boolean, required: false

      field :review, Beta::Types::Objects::ReviewType, null: true

      def resolve(
        review_id:,
        body:,
        title: nil,
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
        work = work_record.work
        record = work_record.record
        oauth_application = context[:doorkeeper_token].application

        form = Forms::WorkRecordForm.new(user: viewer, record: record, work: work, oauth_application: oauth_application)
        form.attributes = {
          comment: body,
          rating_animation: rating_animation_state,
          rating_character: rating_character_state,
          rating_music: rating_music_state,
          rating_overall: rating_overall_state,
          rating_story: rating_story_state,
          share_to_twitter: share_twitter
        }

        if form.invalid?
          raise GraphQL::ExecutionError, form.errors.full_messages.first
        end

        result = Updaters::WorkRecordUpdater.new(
          user: viewer,
          form: form
        ).call

        {
          review: result.record.work_record
        }
      end
    end
  end
end
