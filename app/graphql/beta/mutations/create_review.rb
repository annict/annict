# frozen_string_literal: true

module Beta
  module Mutations
    class CreateReview < Beta::Mutations::Base
      argument :work_id, ID, required: true
      argument :title, String, required: false
      argument :body, String, required: true
      WorkRecord::STATES.each do |state|
        argument state.to_s.underscore.to_sym, Beta::Types::Enums::RatingState, required: false
      end
      argument :share_twitter, Boolean, required: false
      argument :share_facebook, Boolean, required: false

      field :review, Beta::Types::Objects::ReviewType, null: true

      def resolve( # rubocop:disable Metrics/ParameterLists
        work_id:,
        body:, title: nil,
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
        work = Work.only_kept.find_by_graphql_id(work_id)

        form = Forms::WorkRecordForm.new(
          work: work,
          comment: title.present? ? "#{title}\n\n#{body}" : body,
          oauth_application: context[:doorkeeper_token].application,
          rating_animation: rating_animation_state,
          rating_character: rating_character_state,
          rating_music: rating_music_state,
          rating_overall: rating_overall_state,
          rating_story: rating_story_state,
          share_to_twitter: share_twitter&.to_s
        )
        form.user = viewer

        if form.invalid?
          raise GraphQL::ExecutionError, form.errors.full_messages.first
        end

        result = Creators::WorkRecordCreator.new(
          user: viewer,
          form: form
        ).call

        {
          review: viewer.work_records.find_by!(record_id: result.record.id)
        }
      end
    end
  end
end
