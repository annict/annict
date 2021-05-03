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

        result = CreateAnimeRecordService.new(
          user: viewer,
          anime: work,
          rating_overall: rating_overall_state,
          rating_animation: rating_animation_state,
          rating_music: rating_music_state,
          rating_story: rating_story_state,
          rating_character: rating_character_state,
          comment: title.present? ? "#{title}\n\n#{body}" : body,
          share_to_twitter: share_twitter&.to_s
        ).call

        unless result.success?
          raise GraphQL::ExecutionError, result.errors.first.message
        end

        {
          review: viewer.work_records.find_by!(record_id: result.record.id)
        }
      end
    end
  end
end
