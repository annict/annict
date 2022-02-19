# frozen_string_literal: true

module Canary
  module Mutations
    class CreateWorkRecord < Canary::Mutations::Base
      argument :work_id, ID,
        required: true
      argument :rating_overall, Canary::Types::Enums::Rating,
        required: false,
        description: "アニメの評価 (全体)"
      argument :rating_animation, Canary::Types::Enums::Rating,
        required: false,
        description: "アニメの評価 (映像)"
      argument :rating_music, Canary::Types::Enums::Rating,
        required: false,
        description: "アニメの評価 (音楽)"
      argument :rating_story, Canary::Types::Enums::Rating,
        required: false,
        description: "アニメの評価 (ストーリー)"
      argument :rating_character, Canary::Types::Enums::Rating,
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
        work_id:,
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
        work = Work.only_kept.find_by_graphql_id(work_id)

        form = Forms::WorkRecordForm.new(user: viewer, work: work)
        form.attributes = {
          rating_overall: rating_overall,
          rating_animation: rating_animation,
          rating_music: rating_music,
          rating_story: rating_story,
          rating_character: rating_character,
          comment: comment,
          share_to_twitter: share_to_twitter
        }

        if form.invalid?
          return {
            record: nil,
            errors: form.errors.full_messages.map { |message| {message: message} }
          }
        end

        result = Creators::WorkRecordCreator.new(user: viewer, form: form).call

        {
          record: result.record,
          errors: []
        }
      end
    end
  end
end
