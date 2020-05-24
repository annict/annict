# frozen_string_literal: true

module Mutations
  class CreateReview < Mutations::Base
    include V4::GraphqlRunnable

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
      raise Annict::Errors::InvalidAPITokenScopeError unless context[:doorkeeper_token].writable?


      viewer = context[:viewer]
      work = Work.only_kept.find_by_graphql_id(work_id)

      body = title.present? ? "#{title}\n\n#{body}" : body
      work_record_params = {
        body: body,
        rating_animation_state: rating_animation_state,
        rating_music_state: rating_music_state,
        rating_story_state: rating_story_state,
        rating_character_state: rating_character_state,
        rating_overall_state: rating_overall_state,
        share_to_twitter: share_twitter&.to_s
      }

      work_record_entity, err = CreateWorkRecordRepository.
        new(graphql_client: graphql_client(viewer: viewer)).
        create(work: work, params: work_record_params)

      if err
        raise GraphQL::ExecutionError, err.message
      end

      {
        review: viewer.work_records.find(work_record_entity.id)
      }
    end
  end
end
