# frozen_string_literal: true

module Beta
  module Mutations
    class DeleteRecord < Beta::Mutations::Base
      argument :record_id, ID, required: true

      field :episode, Beta::Types::Objects::EpisodeType, null: true

      def resolve(record_id:)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:doorkeeper_token].writable?

        viewer = context[:viewer]
        type_name, item_id = Beta::AnnictSchema.decode_id(record_id)
        episode_record = Object.const_get(type_name).eager_load(:record).merge(viewer.records.only_kept).find(item_id)
        record = episode_record.record

        Destroyers::RecordDestroyer.new(record: record).call

        {
          episode: record.episode
        }
      end
    end
  end
end
