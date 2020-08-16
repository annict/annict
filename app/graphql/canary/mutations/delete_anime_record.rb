# frozen_string_literal: true

module Canary
  module Mutations
    class DeleteAnimeRecord < Canary::Mutations::Base
      argument :anime_record_id, ID, required: true

      field :anime, Canary::Types::Objects::AnimeType, null: true

      def resolve(anime_record_id:)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        work_record = context[:viewer].work_records.only_kept.find_by_graphql_id(anime_record_id)
        work_record.record.destroy

        {
          anime: work_record.work
        }
      end
    end
  end
end
