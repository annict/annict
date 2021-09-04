# frozen_string_literal: true

module Beta
  module Mutations
    class UpdateReview < Beta::Mutations::Base
      argument :review_id, ID, required: true
      argument :title, String, required: false
      argument :body, String, required: true
      argument :rating_overall_state, Beta::Types::Enums::RatingState, required: true
      argument :rating_animation_state, Beta::Types::Enums::RatingState, required: true
      argument :rating_music_state, Beta::Types::Enums::RatingState, required: true
      argument :rating_story_state, Beta::Types::Enums::RatingState, required: true
      argument :rating_character_state, Beta::Types::Enums::RatingState, required: true
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
        work_record = WorkRecord.eager_load(:record).merge(context[:viewer].records.only_kept).find_by_graphql_id(review_id)
        record = work_record.record
        work = record.work

        form = Forms::WorkRecordForm.new(
          work: work,
          deprecated_title: title,
          body: body,
          oauth_application: context[:doorkeeper_token].application,
          rating: rating_overall_state,
          animation_rating: rating_animation_state,
          character_rating: rating_character_state,
          music_rating: rating_music_state,
          story_rating: rating_story_state,
          record: record,
          share_to_twitter: share_twitter
        )

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
