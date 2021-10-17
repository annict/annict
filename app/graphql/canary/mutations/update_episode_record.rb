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
        episode = record.episode_record.episode

        unless record.episode_record?
          raise GraphQL::ExecutionError, "record_id #{record_id} is not an episode record"
        end

        form = Forms::EpisodeRecordForm.new(user: viewer, record: record, episode: episode, oauth_application: context[:application])
        form.attributes = {
          comment: comment,
          rating: rating,
          share_to_twitter: share_to_twitter
        }

        if form.invalid?
          return {
            record: nil,
            errors: form.errors.full_messages.map { |message| {message: message} }
          }
        end

        result = Updaters::EpisodeRecordUpdater.new(
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
