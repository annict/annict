# frozen_string_literal: true

module Canary
  module Mutations
    class UpdateEpisodeRecord < Canary::Mutations::Base
      argument :record_id, ID,
        required: true
      argument :rating, Canary::Types::Enums::Rating,
        required: false,
        description: "エピソードの評価"
      argument :comment, String,
        required: false,
        description: "エピソードの感想"
      argument :share_to_twitter, Boolean,
        required: false,
        description: "記録をTwitterでシェアするかどうか"

      field :record, Canary::Types::Objects::RecordType, null: true
      field :errors, [Canary::Types::Objects::ClientErrorType], null: false

      def resolve(record_id:, rating: nil, comment: nil, share_to_twitter: nil)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        viewer = context[:viewer]
        record = viewer.records.only_kept.find_by_graphql_id(record_id)

        unless record.episode_record?
          raise GraphQL::ExecutionError, "record_id #{record_id} is not an episode record"
        end

        result = UpdateEpisodeRecordService.new(
          user: viewer,
          record: record,
          rating: rating,
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
