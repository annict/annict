# typed: false
# frozen_string_literal: true

module Canary
  module Mutations
    class BulkCreateEpisodeRecords < Canary::Mutations::Base
      argument :episode_ids, [ID],
        required: true

      field :bulk_operation, Canary::Types::Objects::BulkOperationType, null: false

      def resolve(episode_ids:)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        viewer = context[:viewer]
        episode_database_ids = episode_ids.map { |episode_id| GraphQL::Schema::UniqueWithinType.decode(episode_id)[1] }.compact

        job = BulkCreateEpisodeRecordsJob.perform_later(viewer.id, episode_database_ids)
        bulk_operation = OpenStruct.new(job_id: job.provider_job_id)

        {
          bulk_operation: bulk_operation
        }
      end
    end
  end
end
