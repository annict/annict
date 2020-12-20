# frozen_string_literal: true

module Canary
  module Mutations
    class CreateAnimeRecord < Canary::Mutations::Base
      argument :anime_id, ID,
        required: true
      WorkRecord::RATING_FIELDS.each do |rating_field|
        argument rating_field.to_s.camelcase(:lower).to_sym, Canary::Types::Enums::RatingState,
          required: false,
          description: "作品への評価"
      end
      argument :comment, String,
        required: false,
        description: "作品への感想"
      argument :share_to_twitter, Boolean,
        required: false,
        description: "作品への記録をTwitterでシェアするかどうか"

      field :record, Canary::Types::Objects::RecordType, null: true
      field :errors, [Canary::Types::Objects::ClientErrorType], null: false

      def resolve( # rubocop:disable Metrics/ParameterLists
        anime_id:,
        rating_overall: nil,
        rating_animation: nil,
        rating_music: nil,
        rating_story: nil,
        rating_character: nil,
        comment: nil,
        share_to_twitter: nil
      )
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        viewer = context[:viewer]
        anime = Work.only_kept.find_by_graphql_id(anime_id)

        result = CreateAnimeRecordService.new(
          user: viewer,
          anime: anime,
          rating_overall: rating_overall,
          rating_animation: rating_animation,
          rating_music: rating_music,
          rating_story: rating_story,
          rating_character: rating_character,
          comment: comment,
          share_to_twitter: share_to_twitter
        )

        {
          record: result.record,
          errors: result.errors
        }
      end
    end
  end
end
