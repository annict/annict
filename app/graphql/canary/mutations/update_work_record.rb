# frozen_string_literal: true

module Canary
  module Mutations
    class UpdateWorkRecord < Canary::Mutations::Base
      argument :record_id, ID,
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
        work = record.work

        unless record.work_record?
          raise GraphQL::ExecutionError, "record_id #{record_id} is not a work record"
        end

        form = Forms::WorkRecordForm.new(user: viewer, record: record, work: work, oauth_application: context[:application])
        form.attributes = {
          comment: comment,
          rating_animation: rating_animation,
          rating_character: rating_character,
          rating_music: rating_music,
          rating_overall: rating_overall,
          rating_story: rating_story,
          share_to_twitter: share_to_twitter
        }

        if form.invalid?
          return {
            record: nil,
            errors: form.errors.full_messages.map { |message| {message: message} }
          }
        end

        result = Updaters::WorkRecordUpdater.new(
          user: viewer,
          form: form
        ).call

        {
          record: result.record,
          errors: []
        }
      end
    end
  end
end
