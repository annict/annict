# frozen_string_literal: true

module Mutations
  class DeleteRecord < Mutations::Base
    argument :record_id, ID, required: true

    field :episode, Types::Objects::EpisodeType, null: true

    def resolve(record_id:)
      raise Annict::Errors::InvalidAPITokenScopeError unless context[:doorkeeper_token].writable?

      episode_record = context[:viewer].episode_records.without_deleted.find_by_graphql_id(record_id)
      episode_record.record.destroy

      {
        episode: episode_record.episode
      }
    end
  end
end
