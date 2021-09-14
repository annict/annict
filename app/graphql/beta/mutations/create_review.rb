# frozen_string_literal: true

module Beta
  module Mutations
    class CreateReview < Beta::Mutations::Base
      argument :work_id, ID, required: true
      argument :title, String, required: false
      argument :body, String, required: true
      argument :rating_overall_state, Beta::Types::Enums::RatingState, required: false
      argument :rating_animation_state, Beta::Types::Enums::RatingState, required: false
      argument :rating_music_state, Beta::Types::Enums::RatingState, required: false
      argument :rating_story_state, Beta::Types::Enums::RatingState, required: false
      argument :rating_character_state, Beta::Types::Enums::RatingState, required: false
      argument :share_twitter, Boolean, required: false
      argument :share_facebook, Boolean, required: false

      field :review, Beta::Types::Objects::ReviewType, null: true

      def resolve( # rubocop:disable Metrics/ParameterLists
        work_id:,
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
        oauth_application = context[:doorkeeper_token].application
        work = Work.only_kept.find_by_graphql_id(work_id)

        form = Forms::WorkRecordForm.new(user: viewer, work: work, oauth_application: oauth_application)
        form.attributes = {
          deprecated_title: title,
          body: body,
          rating: rating_overall_state,
          animation_rating: rating_animation_state,
          character_rating: rating_character_state,
          music_rating: rating_music_state,
          story_rating: rating_story_state,
          share_to_twitter: share_twitter
        }

        if form.invalid?
          raise GraphQL::ExecutionError, form.errors.full_messages.first
        end

        result = Creators::WorkRecordCreator.new(
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
