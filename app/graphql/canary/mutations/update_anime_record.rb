# frozen_string_literal: true

module Canary
  module Mutations
    class UpdateAnimeRecord < Canary::Mutations::Base
      argument :record_id, ID,
        required: true
      argument :rating_overall, Canary::Types::Enums::RatingState,
        required: false,
        description: "アニメの評価 (全体)"
      argument :rating_animation, Canary::Types::Enums::RatingState,
        required: false,
        description: "アニメの評価 (映像)"
      argument :rating_music, Canary::Types::Enums::RatingState,
        required: false,
        description: "アニメの評価 (音楽)"
      argument :rating_story, Canary::Types::Enums::RatingState,
        required: false,
        description: "アニメの評価 (ストーリー)"
      argument :rating_character, Canary::Types::Enums::RatingState,
        required: false,
        description: "アニメの評価 (キャラクター)"
      argument :comment, String,
        required: false,
        description: "アニメの感想"
      argument :share_to_twitter, Boolean,
        required: false,
        description: "記録をTwitterでシェアするかどうか"

      field :record, Canary::Types::Objects::RecordType, null: true
      field :errors, [Canary::Types::Objects::ClientErrorType], null: false

      def resolve(
        record_id:,
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
        record = viewer.records.only_kept.find_by_graphql_id(record_id)

        unless record.anime_record?
          raise GraphQL::ExecutionError, "record_id #{record_id} is not an anime record"
        end

        result = UpdateAnimeRecordService.new(
          user: viewer,
          record: record,
          rating_overall: rating_overall,
          rating_animation: rating_animation,
          rating_music: rating_music,
          rating_story: rating_story,
          rating_character: rating_character,
          comment: comment,
          share_to_twitter: share_to_twitter,
          oauth_application: context[:application]
        ).call

        {
          record: result.record,
          errors: result.errors
        }
      end
    end
  end
end
